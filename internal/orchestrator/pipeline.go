package orchestrator

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"xpert/internal/agents"
	contextutil "xpert/internal/context"
	"xpert/internal/formatter"
	"xpert/internal/planner"
	"xpert/internal/review"
	"xpert/internal/storage"
	"xpert/internal/structure"
	"xpert/internal/synthesis"
	"xpert/internal/workerpool"
)

type Pipeline struct {
	intentDetector      *planner.IntentDetector
	documentClassifier  *planner.DocumentClassifier
	masterPlanner       *planner.MasterPlanner
	sectionPlanner      *planner.SectionPlanner
	agentFactory        *agents.AgentFactory
	reviewer            *review.Reviewer
	gapDetector         *review.GapDetector
	sectionSynthesizer  *synthesis.SectionSynthesizer
	globalSynthesizer   *synthesis.GlobalSynthesizer
	documentStructurer  *structure.DocumentStructurer
	formatterSet        *formatter.Set
	maxParallelSections int
	plannerWorkers      int
	agentWorkers        int
	reviewWorkers       int
	synthesisWorkers    int
}

type ProgressReporter func(stage storage.JobStage, progress int)

func NewPipeline(
	intentDetector *planner.IntentDetector,
	documentClassifier *planner.DocumentClassifier,
	masterPlanner *planner.MasterPlanner,
	sectionPlanner *planner.SectionPlanner,
	agentFactory *agents.AgentFactory,
	reviewer *review.Reviewer,
	gapDetector *review.GapDetector,
	sectionSynth *synthesis.SectionSynthesizer,
	globalSynth *synthesis.GlobalSynthesizer,
	documentStructurer *structure.DocumentStructurer,
	formatterSet *formatter.Set,
	maxParallelSections int,
) *Pipeline {
	if maxParallelSections < 1 {
		maxParallelSections = 1
	}
	plannerWorkers := maxParallelSections
	reviewWorkers := maxParallelSections
	synthesisWorkers := maxParallelSections
	agentWorkers := maxParallelSections * 4
	if agentWorkers < runtime.NumCPU()*4 {
		agentWorkers = runtime.NumCPU() * 4
	}
	return &Pipeline{
		intentDetector:      intentDetector,
		documentClassifier:  documentClassifier,
		masterPlanner:       masterPlanner,
		sectionPlanner:      sectionPlanner,
		agentFactory:        agentFactory,
		reviewer:            reviewer,
		gapDetector:         gapDetector,
		sectionSynthesizer:  sectionSynth,
		globalSynthesizer:   globalSynth,
		documentStructurer:  documentStructurer,
		formatterSet:        formatterSet,
		maxParallelSections: maxParallelSections,
		plannerWorkers:      plannerWorkers,
		agentWorkers:        agentWorkers,
		reviewWorkers:       reviewWorkers,
		synthesisWorkers:    synthesisWorkers,
	}
}

func (p *Pipeline) Run(request storage.DocumentRequest, report ProgressReporter) (string, string, storage.PipelineTrace, error) {
	// Reset model usage tracker for this run
	p.agentFactory.ResetTracker()

	intent := p.intentDetector.Detect(request.Prompt)
	documentType, focusTerms := p.documentClassifier.Classify(request.DocumentType, request.Prompt)
	request.DocumentType = documentType

	sections := p.masterPlanner.Plan(request)
	plannerPool := workerpool.New(p.plannerWorkers, p.plannerWorkers*2)
	defer plannerPool.Stop()
	var plannerWG sync.WaitGroup
	for i := range sections {
		index := i
		plannerWG.Add(1)
		plannerPool.Submit(func() {
			defer plannerWG.Done()
			sections[index] = p.sectionPlanner.Expand(request, sections[index])
			sections[index] = p.documentStructurer.EnrichSectionPlan(documentType, sections[index])
		})
	}
	plannerWG.Wait()
	if report != nil {
		report(storage.StageGenerating, 30)
	}

	type sectionResult struct {
		Title       string
		Content     string
		ReviewNotes []string
		Summary     string
		Terms       []string
		Err         error
	}

	agentPool := workerpool.New(p.agentWorkers, p.agentWorkers*2)
	defer agentPool.Stop()

	results := make([]sectionResult, len(sections))
	partsBySection := make([][]string, len(sections))
	var generationWG sync.WaitGroup
	var generationErr error
	var generationErrMu sync.Mutex
	setGenerationErr := func(err error) {
		if err == nil {
			return
		}
		generationErrMu.Lock()
		if generationErr == nil {
			generationErr = err
		}
		generationErrMu.Unlock()
	}

	for i, section := range sections {
		partsBySection[i] = make([]string, len(section.SubsectionTitles))
		for j, subsection := range section.SubsectionTitles {
			sectionIndex := i
			subIndex := j
			plan := section
			subTitle := subsection
			generationWG.Add(1)
			agentPool.Submit(func() {
				defer generationWG.Done()
				agent := p.agentFactory.ForSubsection(plan, subTitle)
				generated, err := agent.WriteSubsection(request, plan, subTitle)
				if err != nil {
					setGenerationErr(errors.Join(fmt.Errorf("subsection %q failed", subTitle), err))
					return
				}
				partsBySection[sectionIndex][subIndex] = "### " + subTitle + "\n\n" + generated
			})
		}
	}

	generationWG.Wait()
	if generationErr != nil {
		return "", "", storage.PipelineTrace{}, generationErr
	}
	if report != nil {
		report(storage.StageReviewing, 55)
	}

	reviewPool := workerpool.New(p.reviewWorkers, p.reviewWorkers*2)
	defer reviewPool.Stop()
	var reviewWG sync.WaitGroup
	var reviewErr error
	var reviewErrMu sync.Mutex
	setReviewErr := func(err error) {
		if err == nil {
			return
		}
		reviewErrMu.Lock()
		if reviewErr == nil {
			reviewErr = err
		}
		reviewErrMu.Unlock()
	}

	for i, section := range sections {
		sectionIndex := i
		plan := section
		reviewWG.Add(1)
		reviewPool.Submit(func() {
			defer reviewWG.Done()
			content := strings.Join(partsBySection[sectionIndex], "\n\n")
			reviewNotes := p.reviewer.Review(content)
			if len(reviewNotes) > 0 && reviewNotes[0] != "Section meets the minimum structure requirements." {
				reviser := p.agentFactory.ForSectionRevision(plan)
				revised, err := reviser.ReviseSection(request, plan, content, reviewNotes)
				if err != nil {
					setReviewErr(errors.Join(fmt.Errorf("section %q revision failed", plan.Title), err))
					return
				}
				content = revised
				reviewNotes = p.reviewer.Review(content)
			}
			results[sectionIndex] = sectionResult{
				Title:       plan.Title,
				Content:     content,
				ReviewNotes: reviewNotes,
				Summary:     contextutil.SummarizeText(content, 2),
				Terms:       contextutil.ExtractTerminology(plan.Title, plan.Objective, content),
			}
		})
	}

	reviewWG.Wait()
	if reviewErr != nil {
		return "", "", storage.PipelineTrace{}, reviewErr
	}
	if report != nil {
		report(storage.StageSynthesizing, 75)
	}

	notesBySection := map[string][]string{}
	contentBySection := map[string]string{}
	summariesBySection := map[string]string{}
	sectionDocs := make([]string, len(results))
	terminology := append([]string{}, focusTerms...)
	synthesisPool := workerpool.New(p.synthesisWorkers, p.synthesisWorkers*2)
	defer synthesisPool.Stop()
	var synthesisWG sync.WaitGroup
	for i, result := range results {
		index := i
		sectionResult := result
		if sectionResult.Err != nil {
			return "", "", storage.PipelineTrace{}, errors.Join(fmt.Errorf("section %q failed", sectionResult.Title), sectionResult.Err)
		}
		notesBySection[sectionResult.Title] = sectionResult.ReviewNotes
		contentBySection[sectionResult.Title] = sectionResult.Content
		summariesBySection[sectionResult.Title] = sectionResult.Summary
		terminology = append(terminology, sectionResult.Terms...)
		synthesisWG.Add(1)
		synthesisPool.Submit(func() {
			defer synthesisWG.Done()
			sectionDocs[index] = p.sectionSynthesizer.Synthesize(sectionResult.Title, sectionResult.Content, sectionResult.ReviewNotes)
		})
	}

	synthesisWG.Wait()
	if report != nil {
		report(storage.StageFormatting, 90)
	}
	reviewSummary := p.gapDetector.Summarize(notesBySection, contentBySection)
	title := fmt.Sprintf("%s: %s", request.DocumentType, truncate(request.Prompt, 80))
	markdown := p.globalSynthesizer.Assemble(title, sectionDocs, reviewSummary)
	markdown = p.documentStructurer.Apply(documentType, title, markdown)
	formatted := p.formatterSet.Format(strings.ToLower(request.OutputFormat), title, markdown, reviewSummary)

	// Capture model usage stats
	modelsUsed, fallbackCount := p.agentFactory.GetModelUsageStats()

	return markdown, formatted, storage.PipelineTrace{
		Intent:             intent,
		DocumentType:       documentType,
		DocumentFocus:      firstOr(documentType, focusTerms),
		GlobalContext:      contextutil.CompressTerms(focusTerms, 8),
		Terminology:        contextutil.CompressTerms(uniqueTerms(terminology), 16),
		SectionSummaries:   summariesBySection,
		StructureChecklist: p.documentStructurer.Checklist(documentType),
		Sections:           sections,
		ReviewSummary:      reviewSummary,
		ModelsUsed:         modelsUsed,
		ModelFallbacks:     fallbackCount,
	}, nil
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}

func firstOr(fallback string, values []string) string {
	if len(values) == 0 {
		return fallback
	}
	return values[0]
}

func uniqueTerms(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = contextutil.NormalizeWhitespace(strings.ToLower(value))
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
