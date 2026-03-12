package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"xpert/internal/storage"
)

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>xPert Dashboard</title>
  <style>
    @import url('https://fonts.googleapis.com/css2?family=Fraunces:wght@600;700&family=Space+Grotesk:wght@400;500;600&display=swap');
    :root{
      --bg:#f7f2ea;
      --ink:#19212a;
      --muted:#5f6b7a;
      --brand:#1f4b7a;
      --accent:#d97706;
      --card:#ffffff;
      --line:#e0d8cc;
      --shadow:0 20px 50px rgba(18, 31, 46, 0.12);
    }
    *{box-sizing:border-box}
    body{
      margin:0;
      font-family:'Space Grotesk', sans-serif;
      color:var(--ink);
      background:
        radial-gradient(1200px 600px at 10% -10%, rgba(217, 119, 6, 0.18), transparent 60%),
        radial-gradient(1000px 700px at 110% 10%, rgba(31, 75, 122, 0.18), transparent 55%),
        var(--bg);
    }
    .shell{
      max-width:1200px;
      margin:0 auto;
      padding:48px 22px 72px;
    }
    header{
      display:flex;
      gap:24px;
      align-items:flex-end;
      justify-content:space-between;
      flex-wrap:wrap;
      margin-bottom:28px;
      animation:fadeUp 700ms ease-out both;
    }
    h1{
      font-family:'Fraunces', serif;
      font-size:40px;
      margin:0 0 8px;
      letter-spacing:0.3px;
    }
    .subtitle{
      margin:0;
      color:var(--muted);
      font-size:16px;
    }
    .meta{
      display:flex;
      gap:16px;
      flex-wrap:wrap;
      color:var(--muted);
      font-size:14px;
    }
    .pill{
      background:rgba(255,255,255,0.7);
      border:1px solid var(--line);
      border-radius:999px;
      padding:6px 12px;
    }
    .grid{
      display:grid;
      grid-template-columns:minmax(0, 2fr) minmax(0, 1fr);
      gap:22px;
      align-items:start;
    }
    .card{
      background:var(--card);
      border:1px solid var(--line);
      border-radius:18px;
      padding:22px;
      box-shadow:var(--shadow);
      animation:fadeUp 700ms ease-out both;
    }
    .card h2{
      font-family:'Fraunces', serif;
      margin:0 0 12px;
      font-size:22px;
    }
    label{
      display:block;
      font-weight:600;
      margin-bottom:6px;
      color:var(--ink);
    }
    input, select, textarea{
      width:100%;
      border:1px solid var(--line);
      border-radius:12px;
      padding:12px 14px;
      font-size:14px;
      font-family:inherit;
      background:#fff;
      color:var(--ink);
      transition:border 150ms ease, box-shadow 150ms ease;
    }
    input:focus, select:focus, textarea:focus{
      outline:none;
      border-color:var(--brand);
      box-shadow:0 0 0 3px rgba(31, 75, 122, 0.15);
    }
    textarea{min-height:140px;resize:vertical;}
    .row{display:grid;grid-template-columns:repeat(2, minmax(0,1fr));gap:14px;}
    .row + .row, .field + .field{margin-top:14px;}
    .actions{display:flex;gap:12px;align-items:center;margin-top:18px;}
    button{
      border:none;
      border-radius:12px;
      padding:12px 18px;
      font-weight:600;
      font-size:14px;
      cursor:pointer;
      background:var(--brand);
      color:#fff;
      box-shadow:0 12px 24px rgba(31,75,122,0.18);
      transition:transform 150ms ease, box-shadow 150ms ease;
    }
    button:hover{transform:translateY(-1px);box-shadow:0 16px 28px rgba(31,75,122,0.22);}
    .secondary{
      background:transparent;
      color:var(--brand);
      border:1px solid var(--brand);
      box-shadow:none;
    }
    .note{
      font-size:13px;
      color:var(--muted);
      margin:10px 0 0;
    }
    .status{
      background:#f4f6f8;
      border-radius:12px;
      padding:14px;
      font-size:13px;
      color:#2d3640;
      border:1px dashed #c9d0d9;
      min-height:74px;
    }
    .steps li{margin-bottom:8px;color:var(--muted);}
    .tiny{
      font-size:12px;
      color:var(--muted);
    }
    .endpoint{font-family:monospace;font-size:12px;background:#f3ede5;padding:2px 6px;border-radius:6px;}
    @keyframes fadeUp{
      from{opacity:0;transform:translateY(12px);}to{opacity:1;transform:translateY(0);}
    }
    @media (max-width: 980px){
      .grid{grid-template-columns:1fr;}
      .row{grid-template-columns:1fr;}
    }
  </style>
</head>
<body>
  <div class="shell">
    <header>
      <div>
        <h1>xPert Dashboard</h1>
        <p class="subtitle">Submit a document request and track generation progress.</p>
      </div>
      <div class="meta">
        <span class="pill">Storage: <strong>` + s.config.StorageBackend + `</strong></span>
        <span class="pill">LLM: <strong>` + s.config.LLMProvider + `</strong></span>
      </div>
    </header>

    <div class="grid">
      <section class="card">
        <h2>New Request</h2>
        <form id="request-form">
          <div class="field">
            <label for="prompt">Question / Request</label>
            <textarea id="prompt" name="prompt" placeholder="Describe the document you need. Example: Generate a detailed SOP for Web Penetration Testing covering scope, tools, and validation." required></textarea>
          </div>
          <div class="row">
            <div class="field">
              <label for="document_type">Document Type</label>
              <input id="document_type" name="document_type" placeholder="SOP, report, guide, playbook" />
            </div>
            <div class="field">
              <label for="tone">Tone</label>
              <select id="tone" name="tone">
                <option value="">Auto</option>
                <option>Executive</option>
                <option>Technical</option>
                <option>Neutral</option>
                <option>Conversational</option>
              </select>
            </div>
          </div>
          <div class="row">
            <div class="field">
              <label for="output_format">Output Format</label>
              <select id="output_format" name="output_format">
                <option value="markdown">Markdown</option>
                <option value="html">HTML</option>
                <option value="pdf">PDF</option>
              </select>
            </div>
            <div class="field">
              <label for="target_word_count">Target Word Count</label>
              <input id="target_word_count" name="target_word_count" type="number" min="500" step="250" value="6000" />
            </div>
          </div>
          <div class="actions">
            <button type="submit">Create Job</button>
            <button class="secondary" type="button" id="reset-btn">Clear</button>
          </div>
          <p class="note">Jobs are created via <span class="endpoint">POST /documents</span>. Use the returned job ID to fetch results.</p>
        </form>
        <div class="field" style="margin-top:18px;">
          <label>Latest Status</label>
          <div id="status" class="status">No request submitted yet.</div>
        </div>
        <div class="field" style="margin-top:14px;">
          <label for="lookup">Lookup Job or Document ID</label>
          <div class="row">
            <input id="lookup" placeholder="Job ID or Document ID" />
            <button class="secondary" type="button" id="lookup-btn">Fetch</button>
          </div>
          <p class="tiny">Uses <span class="endpoint">GET /jobs/{id}</span> or <span class="endpoint">GET /documents/{id}</span>.</p>
        </div>
      </section>

      <aside class="card">
        <h2>Pipeline Flow</h2>
        <ol class="steps">
          <li>Intent Detector</li>
          <li>Document Classifier</li>
          <li>Master Planner</li>
          <li>Section Planners</li>
          <li>Dynamic Expert Agents</li>
          <li>Section Reviewers</li>
          <li>Gap Detector</li>
          <li>Global Synthesizer</li>
          <li>Document Structurer</li>
          <li>Formatter</li>
        </ol>
        <h2 style="margin-top:22px;">Endpoints</h2>
        <div class="tiny">
          <div><span class="endpoint">POST /documents</span> create job</div>
          <div><span class="endpoint">POST /documents/batch</span> create batch</div>
          <div><span class="endpoint">GET /jobs/{id}</span> job detail</div>
          <div><span class="endpoint">GET /documents/{id}</span> document detail</div>
        </div>
      </aside>
    </div>
  </div>

  <script>
    const form = document.getElementById('request-form');
    const statusBox = document.getElementById('status');
    const resetBtn = document.getElementById('reset-btn');
    const lookupBtn = document.getElementById('lookup-btn');
    const lookupInput = document.getElementById('lookup');

    const setStatus = (message) => {
      statusBox.textContent = message;
    };

    form.addEventListener('submit', async (event) => {
      event.preventDefault();
      const payload = {
        prompt: document.getElementById('prompt').value.trim(),
        document_type: document.getElementById('document_type').value.trim(),
        tone: document.getElementById('tone').value,
        output_format: document.getElementById('output_format').value,
        target_word_count: Number(document.getElementById('target_word_count').value || 0),
      };
      setStatus('Submitting job...');
      try {
        const response = await fetch('/documents', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload),
        });
        const data = await response.json();
        if (!response.ok) {
          setStatus('Request failed: ' + (data.error || response.statusText));
          return;
        }
        const jobId = data.id || data.ID;
        const docId = data.document_id || data.DocumentID;
        setStatus('Job created. Job ID: ' + jobId + '. Document ID: ' + docId + '.');
      } catch (err) {
        setStatus('Request failed: ' + err.message);
      }
    });

    resetBtn.addEventListener('click', () => {
      form.reset();
      setStatus('Form cleared.');
    });

    lookupBtn.addEventListener('click', async () => {
      const id = lookupInput.value.trim();
      if (!id) {
        setStatus('Enter a job or document ID to fetch.');
        return;
      }
      setStatus('Fetching status...');
      try {
        let response = await fetch('/jobs/' + id);
        if (!response.ok) {
          response = await fetch('/documents/' + id);
        }
        const data = await response.json();
        if (!response.ok) {
          setStatus('Lookup failed: ' + (data.error || response.statusText));
          return;
        }
        setStatus(JSON.stringify(data, null, 2));
      } catch (err) {
        setStatus('Lookup failed: ' + err.message);
      }
    });
  </script>
</body>
</html>`))
}

func (s *Server) handleDocuments(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var payload storage.DocumentRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}
		job, err := s.jobManager.CreateJob(payload)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusAccepted, storage.CreateJobResponse{ID: job.ID, Status: job.Status, DocumentID: job.DocumentID})
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{"documents": s.repository.ListDocuments(50)})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleBatchDocuments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var payload storage.BatchDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	batch, err := s.jobManager.CreateBatchJob(payload)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{
		"id":            batch.ID,
		"status":        batch.Status,
		"child_job_ids": batch.ChildJobIDs,
	})
}

func (s *Server) handleDocumentByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/documents/")
	if id == "" || id == "/documents" {
		http.NotFound(w, r)
		return
	}
	if detail, ok := s.jobManager.GetDocumentDetail(id); ok {
		writeJSON(w, http.StatusOK, detail)
		return
	}
	if detail, ok := s.jobManager.GetJobDetail(id); ok {
		writeJSON(w, http.StatusOK, detail)
		return
	}
	writeJSON(w, http.StatusNotFound, map[string]string{"error": "document not found"})
}

func (s *Server) handleJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"jobs": s.repository.ListJobs(50)})
}

func (s *Server) handleJobByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/jobs/")
	if id == "" {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		if detail, ok := s.jobManager.GetJobDetail(id); ok {
			writeJSON(w, http.StatusOK, detail)
			return
		}
		if detail, ok := s.jobManager.GetBatchDetail(id); ok {
			writeJSON(w, http.StatusOK, detail)
			return
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
	case http.MethodDelete:
		if _, ok := s.jobManager.CancelJob(id); !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
			return
		}
		w.WriteHeader(http.StatusAccepted)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
