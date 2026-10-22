package web

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/maks2134/telegram-rag-bot/internal/domain"
)

// Service runs the RAG ask pipeline.
type Service interface {
	Ask(ctx context.Context, q string) domain.Result
}

// Catalog lists loaded knowledge-base documents.
type Catalog interface {
	Docs() []domain.Document
}

// Server exposes health, ask, docs list, and the Trace UI.
type Server struct {
	Service Service
	Catalog Catalog
	Model   string
}

// Handler returns the HTTP mux.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /docs", s.listDocs)
	mux.HandleFunc("POST /ask", s.ask)
	mux.HandleFunc("GET /", s.index)
	return mux
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]any{"ok": true, "model": s.Model})
}

func (s *Server) listDocs(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, s.Catalog.Docs())
}

func (s *Server) ask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Question string `json:"question"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	req.Question = strings.TrimSpace(req.Question)
	if req.Question == "" {
		http.Error(w, "question required", http.StatusBadRequest)
		return
	}
	res := s.Service.Ask(r.Context(), req.Question)
	writeJSON(w, res)
}

func (s *Server) index(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(indexHTML))
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

const indexHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>DocPilot — Docs → Telegram RAG</title>
  <link rel="preconnect" href="https://fonts.googleapis.com" />
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />
  <link href="https://fonts.googleapis.com/css2?family=Fraunces:opsz,wght@9..144,500;9..144,700&family=Sora:wght@400;500;600&display=swap" rel="stylesheet" />
  <style>
    :root {
      --ink: #14201c;
      --muted: #5a6b64;
      --paper: #f3efe6;
      --panel: rgba(255, 252, 246, 0.82);
      --line: rgba(20, 32, 28, 0.12);
      --accent: #0f7a5f;
      --accent-2: #c45c26;
      --ok: #1f8a5b;
      --skip: #7a857f;
      --err: #b33a3a;
      --shadow: 0 18px 50px rgba(20, 32, 28, 0.08);
      --display: "Fraunces", Georgia, serif;
      --sans: "Sora", system-ui, sans-serif;
      --mono: "SF Mono", ui-monospace, monospace;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0; min-height: 100vh; color: var(--ink); font-family: var(--sans);
      background:
        radial-gradient(900px 520px at 8% -8%, #d8e8df 0%, transparent 55%),
        radial-gradient(700px 420px at 100% 0%, #f0dcc8 0%, transparent 50%),
        linear-gradient(180deg, #efe9dc 0%, var(--paper) 45%, #e7efe9 100%);
      background-attachment: fixed;
    }
    body::before {
      content: ""; position: fixed; inset: 0; pointer-events: none; opacity: 0.35;
      background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.85' numOctaves='3' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)' opacity='0.45'/%3E%3C/svg%3E");
      mix-blend-mode: multiply;
    }
    .wrap { position: relative; max-width: 1120px; margin: 0 auto; padding: 1.75rem 1.25rem 2.5rem; }
    .brand {
      display: flex; flex-wrap: wrap; align-items: flex-end; justify-content: space-between;
      gap: 1rem; margin-bottom: 1.35rem;
      animation: rise 0.7s ease both;
    }
    .brand-mark {
      width: 52px; height: 52px; border-radius: 16px;
      background: linear-gradient(145deg, #1a9b78, #0d5c49);
      box-shadow: 0 10px 24px rgba(15, 122, 95, 0.28);
      display: grid; place-items: center; color: #f6fff9; font-family: var(--display);
      font-weight: 700; font-size: 1.35rem;
    }
    .brand h1 {
      margin: 0; font-family: var(--display); font-size: clamp(2rem, 4vw, 2.75rem);
      letter-spacing: -0.03em; line-height: 1.05; font-weight: 700;
    }
    .brand p { margin: 0.45rem 0 0; color: var(--muted); max-width: 42ch; line-height: 1.45; }
    .pill {
      align-self: center; font-size: 0.78rem; color: var(--accent);
      border: 1px solid color-mix(in srgb, var(--accent) 35%, transparent);
      background: color-mix(in srgb, var(--accent) 8%, white);
      padding: 0.4rem 0.7rem; border-radius: 999px; white-space: nowrap;
    }
    main {
      display: grid; gap: 1rem; grid-template-columns: 1fr;
      animation: rise 0.85s 0.08s ease both;
    }
    @media (min-width: 920px) {
      main { grid-template-columns: 1.15fr 0.85fr; align-items: start; }
    }
    .panel {
      background: var(--panel); border: 1px solid var(--line);
      border-radius: 22px; padding: 1.1rem 1.15rem; box-shadow: var(--shadow);
      backdrop-filter: blur(10px);
    }
    .panel h2 {
      margin: 0 0 0.85rem; font-family: var(--display); font-size: 1.2rem; font-weight: 600;
    }
    .chips { display: flex; flex-wrap: wrap; gap: 0.45rem; margin-bottom: 0.85rem; }
    .chip {
      border: 1px solid var(--line); background: #fffdf8; color: var(--ink);
      border-radius: 999px; padding: 0.45rem 0.75rem; font-size: 0.82rem;
      cursor: pointer; transition: transform 0.15s ease, border-color 0.15s ease, background 0.15s ease;
    }
    .chip:hover { transform: translateY(-1px); border-color: color-mix(in srgb, var(--accent) 45%, var(--line)); background: #f3faf6; }
    label { display: block; font-size: 0.78rem; color: var(--muted); margin-bottom: 0.35rem; font-weight: 500; }
    textarea {
      width: 100%; min-height: 110px; resize: vertical; font: inherit;
      background: #fffdf8; color: var(--ink); border: 1px solid var(--line);
      border-radius: 14px; padding: 0.85rem 0.95rem; outline: none;
      transition: border-color 0.15s ease, box-shadow 0.15s ease;
    }
    textarea:focus {
      border-color: color-mix(in srgb, var(--accent) 55%, var(--line));
      box-shadow: 0 0 0 3px color-mix(in srgb, var(--accent) 16%, transparent);
    }
    .actions { display: flex; flex-wrap: wrap; gap: 0.55rem; margin-top: 0.8rem; }
    .btn {
      border: 0; border-radius: 12px; padding: 0.7rem 1.15rem; font: inherit; font-weight: 600;
      cursor: pointer; transition: transform 0.15s ease, opacity 0.15s ease;
    }
    .btn:hover { transform: translateY(-1px); }
    .btn:disabled { opacity: 0.55; cursor: wait; transform: none; }
    .btn-primary { background: var(--accent); color: #f4fff9; }
    .btn-ghost { background: transparent; color: var(--muted); border: 1px solid var(--line); }
    .answer {
      margin-top: 1rem; white-space: pre-wrap; line-height: 1.55;
      padding: 0.95rem 1rem; border-radius: 14px; background: #fffdf8; border: 1px solid var(--line);
      min-height: 4.5rem; animation: fade 0.35s ease;
    }
    .sources { margin-top: 0.7rem; color: var(--muted); font-size: 0.88rem; }
    .meta { color: var(--muted); font-size: 0.78rem; margin-top: 0.45rem; font-family: var(--mono); }
    .rail {
      display: grid; grid-template-columns: repeat(4, 1fr); gap: 0.35rem; margin-bottom: 0.9rem;
    }
    .rail span {
      text-align: center; font-size: 0.68rem; text-transform: uppercase; letter-spacing: 0.04em;
      padding: 0.4rem 0.2rem; border-radius: 10px; border: 1px dashed var(--line); color: var(--skip);
      transition: background 0.25s ease, color 0.25s ease, border-color 0.25s ease;
    }
    .rail span.on { background: color-mix(in srgb, var(--ok) 12%, white); color: var(--ok); border-style: solid; border-color: color-mix(in srgb, var(--ok) 35%, var(--line)); }
    .rail span.bad { background: color-mix(in srgb, var(--err) 10%, white); color: var(--err); border-style: solid; }
    .step {
      border: 1px solid var(--line); border-radius: 14px; padding: 0.75rem 0.85rem;
      margin-bottom: 0.55rem; background: #fffdf8;
      animation: rise 0.4s ease both;
    }
    .step:nth-child(2) { animation-delay: 0.05s; }
    .step:nth-child(3) { animation-delay: 0.1s; }
    .step:nth-child(4) { animation-delay: 0.15s; }
    .step-head { display: flex; justify-content: space-between; gap: 0.5rem; align-items: center; }
    .step-name { font-family: var(--mono); font-size: 0.82rem; font-weight: 500; }
    .badge {
      font-size: 0.68rem; text-transform: uppercase; letter-spacing: 0.04em;
      padding: 0.18rem 0.45rem; border-radius: 999px; border: 1px solid var(--line);
    }
    .badge.ok { color: var(--ok); border-color: color-mix(in srgb, var(--ok) 40%, var(--line)); }
    .badge.skip { color: var(--skip); }
    .badge.error { color: var(--err); border-color: color-mix(in srgb, var(--err) 40%, var(--line)); }
    .step-detail { margin: 0.35rem 0 0; color: var(--muted); font-size: 0.84rem; }
    .step pre {
      margin: 0.45rem 0 0; padding: 0.55rem; overflow: auto; max-height: 130px;
      font-family: var(--mono); font-size: 0.72rem; background: #f1ebe0; border-radius: 10px;
    }
    .docs {
      margin-top: 0.85rem; display: flex; flex-wrap: wrap; gap: 0.4rem;
    }
    .doc-tag {
      font-size: 0.75rem; color: var(--muted); border: 1px solid var(--line);
      border-radius: 8px; padding: 0.25rem 0.5rem; background: rgba(255,255,255,0.5);
    }
    @keyframes rise {
      from { opacity: 0; transform: translateY(10px); }
      to { opacity: 1; transform: none; }
    }
    @keyframes fade {
      from { opacity: 0; }
      to { opacity: 1; }
    }
  </style>
</head>
<body>
  <div class="wrap">
    <header class="brand">
      <div style="display:flex;gap:0.9rem;align-items:flex-start">
        <div class="brand-mark" aria-hidden="true">D</div>
        <div>
          <h1>DocPilot</h1>
          <p>Grounded answers from your Markdown docs — same RAG pipeline that powers Telegram, with a live retrieve → LLM trace.</p>
        </div>
      </div>
      <div class="pill" id="modelPill">model …</div>
    </header>
    <main>
      <section class="panel">
        <h2>Ask the knowledge base</h2>
        <div class="chips" id="chips">
          <button type="button" class="chip" data-q="How long do refunds take?">Refunds timing</button>
          <button type="button" class="chip" data-q="How does onboarding work for new teams?">Onboarding</button>
          <button type="button" class="chip" data-q="What is the Business plan SLA?">SLA</button>
          <button type="button" class="chip" data-q="Can I get a free unicorn?" data-expect="refuse">Out-of-scope</button>
        </div>
        <label for="q">Question</label>
        <textarea id="q" placeholder="How long do refunds take?"></textarea>
        <div class="actions">
          <button class="btn btn-primary" id="ask">Ask DocPilot</button>
          <button class="btn btn-ghost" id="clear" type="button">Clear</button>
        </div>
        <div class="answer" id="answer">Pick a chip or type a question — the pipeline trace updates on the right.</div>
        <div class="sources" id="sources"></div>
        <div class="meta" id="meta"></div>
        <div class="docs" id="docs"></div>
      </section>
      <section class="panel">
        <h2>Pipeline trace</h2>
        <div class="rail" id="rail">
          <span data-step="retrieve">retrieve</span>
          <span data-step="prompt">prompt</span>
          <span data-step="llm">llm</span>
          <span data-step="answer">answer</span>
        </div>
        <div id="trace"><p class="step-detail">Waiting for a question…</p></div>
      </section>
    </main>
  </div>
  <script>
    const askBtn = document.getElementById('ask');
    const qEl = document.getElementById('q');

    async function boot() {
      try {
        const h = await fetch('/health').then(r => r.json());
        document.getElementById('modelPill').textContent = h.model || 'model ready';
        const docs = await fetch('/docs').then(r => r.json());
        document.getElementById('docs').innerHTML = (docs || []).map(d =>
          '<span class="doc-tag">' + escapeHtml(d.title) + '</span>'
        ).join('');
      } catch (_) {
        document.getElementById('modelPill').textContent = 'offline';
      }
    }

    document.getElementById('chips').onclick = (e) => {
      const btn = e.target.closest('.chip');
      if (!btn) return;
      qEl.value = btn.dataset.q || '';
      ask();
    };
    document.getElementById('clear').onclick = () => {
      qEl.value = '';
      document.getElementById('answer').textContent = 'Pick a chip or type a question — the pipeline trace updates on the right.';
      document.getElementById('sources').textContent = '';
      document.getElementById('meta').textContent = '';
      document.getElementById('trace').innerHTML = '<p class="step-detail">Waiting for a question…</p>';
      resetRail();
    };
    askBtn.onclick = ask;
    qEl.addEventListener('keydown', (e) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') ask();
    });

    function resetRail() {
      document.querySelectorAll('#rail span').forEach(s => s.className = '');
    }

    async function ask() {
      const question = qEl.value.trim();
      if (!question) return;
      askBtn.disabled = true;
      resetRail();
      document.getElementById('answer').textContent = 'Thinking…';
      document.getElementById('sources').textContent = '';
      document.getElementById('meta').textContent = '';
      document.getElementById('trace').innerHTML = '';
      try {
        const res = await fetch('/ask', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ question })
        });
        const data = await res.json();
        if (!res.ok) throw new Error(data.message || res.statusText);
        document.getElementById('answer').textContent = data.answer || '(empty)';
        const src = (data.sources || []).map(s => s.title + ' · score ' + s.score);
        document.getElementById('sources').textContent = src.length
          ? 'Sources: ' + [...new Set((data.sources || []).map(s => s.title))].join(' · ')
          : 'No sources retrieved';
        document.getElementById('meta').textContent =
          (data.model || '?') + ' · ' + (data.latency_ms || 0) + ' ms' + (data.refused ? ' · refused' : '');
        const steps = data.trace || [];
        steps.forEach(step => {
          const el = document.querySelector('#rail [data-step="' + step.name + '"]');
          if (!el) return;
          el.className = step.status === 'error' ? 'bad' : (step.status === 'ok' ? 'on' : '');
        });
        document.getElementById('trace').innerHTML = steps.map(step => {
          const payload = step.payload ? '<pre>' + escapeHtml(JSON.stringify(step.payload, null, 2)) + '</pre>' : '';
          const ms = step.ms != null ? ' · ' + step.ms + ' ms' : '';
          return '<div class="step"><div class="step-head"><span class="step-name">' + escapeHtml(step.name) + '</span>' +
            '<span class="badge ' + escapeHtml(step.status) + '">' + escapeHtml(step.status) + '</span></div>' +
            '<p class="step-detail">' + escapeHtml(step.detail || '') + ms + '</p>' + payload + '</div>';
        }).join('') || '<p class="step-detail">No trace.</p>';
      } catch (e) {
        document.getElementById('answer').textContent = 'Error: ' + e.message;
      } finally {
        askBtn.disabled = false;
      }
    }

    function escapeHtml(s) {
      return String(s).replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
    }
    boot();
  </script>
</body>
</html>
`
