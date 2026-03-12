package orchestrator

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"docgen/internal/agents"
	contextutil "docgen/internal/context"
	"docgen/internal/formatter"
	"docgen/internal/planner"
	"docgen/internal/review"
	"docgen/internal/storage"
	"docgen/internal/structure"
	"docgen/internal/synthesis"
)

type Pipeline struct {
	intentDetector      *planner.IntentDetector
	documentClassifier  *planner.DocumentClassifier
	masterPlanner       *planner.MasterPlanner
	sectionPlanner      *planner.SectionPlanner
	expertAgent         *agents.ExpertAgent
	reviewer            *review.Reviewer
	gapDetector         *review.GapDetector
	sectionSynthesizer  *synthesis.SectionSynthesizer
	globalSynthesizer   *synthesis.GlobalSynthesizer
	documentStructurer  *structure.DocumentStructurer
	formatterSet        *formatter.Set
	maxParallelSections int
}

func NewPipeline(
	intentDetector *planner.IntentDetector,
	documentClassifier *planner.DocumentClassifier,
	masterPlanner *planner.MasterPlanner,
	sectionPlanner *planner.SectionPlanner,
	expertAgent *agents.ExpertAgent,
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
	return &Pipeline{
		intentDetector:      intentDetector,
		documentClassifier:  documentClassifier,
		masterPlanner:       masterPlanner,
		sectionPlanner:      sectionPlanner,
		expertAgent:         expertAgent,
		reviewer:            reviewer,
		gapDetector:         gapDetector,
		sectionSynthesizer:  sectionSynth,
		globalSynthesizer:   globalSynth,
		documentStructurer:  documentStructurer,
		formatterSet:        formatterSet,
		maxParallelSections: maxParallelSections,
	}
}

func (p *Pipeline) Run(request storage.DocumentRequest) (string, string, storage.PipelineTrace, error) {
	intent := p.intentDetector.Detect(request.Prompt)
	documentType, focusTerms := p.documentClassifier.Classify(request.DocumentType, request.Prompt)
	request.DocumentType = documentType

	sections := p.masterPlanner.Plan(request)
	for i := range sections {
		sections[i] = p.sectionPlanner.Expand(request, sections[i])
		sections[i] = p.documentStructurer.EnrichSectionPlan(documentType, sections[i])
	}

	type sectionResult struct {
		Title       string
		Content     string
		ReviewNotes []string
		Summary     string
		Terms       []string
		Err         error
	}

	sem := make(chan struct{}, p.maxParallelSections)
	results := make([]sectionResult, len(sections))
	var wg sync.WaitGroup

	for i, section := range sections {
		wg.Add(1)
		go func(index int, plan storage.SectionPlan) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			parts := make([]string, 0, len(plan.SubsectionTitles))
			for _, subsection := range plan.SubsectionTitles {
				generated, err := p.expertAgent.WriteSubsection(request, plan, subsection)
				if err != nil {
					results[index] = sectionResult{Title: plan.Title, Err: err}
					return
				}
				parts = append(parts, "### "+subsection+"\n\n"+generated)
			}
			content := strings.Join(parts, "\n\n")
			reviewNotes := p.reviewer.Review(content)
			if len(reviewNotes) > 0 && reviewNotes[0] != "Section meets the minimum structure requirements." {
				revised, err := p.expertAgent.ReviseSection(request, plan, content, reviewNotes)
				if err != nil {
					results[index] = sectionResult{Title: plan.Title, Err: err}
					return
				}
				content = revised
				reviewNotes = p.reviewer.Review(content)
			}
			results[index] = sectionResult{
				Title:       plan.Title,
				Content:     content,
				ReviewNotes: reviewNotes,
				Summary:     contextutil.SummarizeText(content, 2),
				Terms:       contextutil.ExtractTerminology(plan.Title, plan.Objective, content),
			}
		}(i, section)
	}

	wg.Wait()

	notesBySection := map[string][]string{}
	contentBySection := map[string]string{}
	summariesBySection := map[string]string{}
	sectionDocs := make([]string, 0, len(results))
	terminology := append([]string{}, focusTerms...)
	for _, result := range results {
		if result.Err != nil {
			return "", "", storage.PipelineTrace{}, errors.Join(fmt.Errorf("section %q failed", result.Title), result.Err)
		}
		notesBySection[result.Title] = result.ReviewNotes
		contentBySection[result.Title] = result.Content
		summariesBySection[result.Title] = result.Summary
		terminology = append(terminology, result.Terms...)
		sectionDocs = append(sectionDocs, p.sectionSynthesizer.Synthesize(result.Title, result.Content, result.ReviewNotes))
	}
	reviewSummary := p.gapDetector.Summarize(notesBySection, contentBySection)
	title := fmt.Sprintf("%s: %s", request.DocumentType, truncate(request.Prompt, 80))
	markdown := p.globalSynthesizer.Assemble(title, sectionDocs, reviewSummary)
	markdown = p.documentStructurer.Apply(documentType, title, markdown)
	formatted := p.formatterSet.Format(strings.ToLower(request.OutputFormat), title, markdown, reviewSummary)

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
