package playwright

import (
	"encoding/json"
	"html/template"
	"io"
	"time"

	"github.com/zikani03/basi/core"
)

// RunReport is the structured summary produced after one or more runs.
type RunReport struct {
	File     string               `json:"file"`
	Passed   bool                 `json:"passed"`
	Runs     int                  `json:"runs"`
	Pass     int                  `json:"pass"`
	Fail     int                  `json:"fail"`
	Steps    []core.ProgressEvent `json:"steps"`
	Duration string               `json:"duration"`
	RunAt    string               `json:"run_at"`
}

// BuildReport constructs a RunReport from collected events.
// Only terminal events (ok/fail) are included in Steps; "running" events are omitted.
func BuildReport(file string, pass, fail int, events []core.ProgressEvent, duration time.Duration) RunReport {
	terminal := make([]core.ProgressEvent, 0, len(events))
	for _, e := range events {
		if e.Status != core.StepRunning {
			terminal = append(terminal, e)
		}
	}
	return RunReport{
		File:     file,
		Passed:   fail == 0,
		Runs:     pass + fail,
		Pass:     pass,
		Fail:     fail,
		Steps:    terminal,
		Duration: duration.String(),
		RunAt:    time.Now().Format(time.RFC3339),
	}
}

// WriteJSONReport writes the report as indented JSON to w.
func WriteJSONReport(w io.Writer, report RunReport) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

// WriteHTMLReport renders an HTML report to w.
func WriteHTMLReport(w io.Writer, report RunReport) error {
	return reportTemplate.Execute(w, report)
}

var reportTemplate = template.Must(template.New("report").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Basi Report — {{.File}}</title>
<style>
  body { font-family: system-ui, sans-serif; margin: 2rem; color: #1a1a1a; }
  h1 { font-size: 1.4rem; margin-bottom: 0.25rem; }
  .meta { color: #666; font-size: 0.85rem; margin-bottom: 1.5rem; }
  .badge { display: inline-block; padding: 0.2rem 0.6rem; border-radius: 4px; font-weight: 600; font-size: 0.8rem; }
  .pass { background: #d4edda; color: #155724; }
  .fail { background: #f8d7da; color: #721c24; }
  .run-badge { background: #d1ecf1; color: #0c5460; }
  table { width: 100%; border-collapse: collapse; font-size: 0.9rem; }
  th { text-align: left; padding: 0.5rem 0.75rem; background: #f5f5f5; border-bottom: 2px solid #ddd; }
  td { padding: 0.45rem 0.75rem; border-bottom: 1px solid #eee; vertical-align: top; }
  tr.ok td { }
  tr.fail td { background: #fff5f5; }
  .status-ok  { color: #28a745; font-weight: 600; }
  .status-fail { color: #dc3545; font-weight: 600; }
  .selector { font-family: monospace; font-size: 0.82rem; color: #555; }
  .msg { color: #c0392b; font-size: 0.82rem; }
  .dur { color: #888; font-size: 0.8rem; }
</style>
</head>
<body>
<h1>Basi Report</h1>
<p class="meta">
  <strong>File:</strong> {{.File}} &nbsp;|&nbsp;
  <strong>Run at:</strong> {{.RunAt}} &nbsp;|&nbsp;
  <strong>Duration:</strong> {{.Duration}} &nbsp;|&nbsp;
  <span class="badge {{if .Passed}}pass{{else}}fail{{end}}">{{if .Passed}}PASSED{{else}}FAILED{{end}}</span>
  {{if gt .Runs 1}}&nbsp;<span class="badge run-badge">{{.Pass}}/{{.Runs}} runs passed</span>{{end}}
</p>
<table>
  <thead>
    <tr>
      <th>#</th>
      {{if gt .Runs 1}}<th>Run</th>{{end}}
      <th>Action</th>
      <th>Selector</th>
      <th>Status</th>
      <th>Duration</th>
      <th>Message</th>
    </tr>
  </thead>
  <tbody>
  {{range .Steps}}
    <tr class="{{.Status}}">
      <td>{{.Step}}</td>
      {{if gt $.Runs 1}}<td>{{.Run}}</td>{{end}}
      <td>{{.Action}}</td>
      <td class="selector">{{.Selector}}</td>
      <td class="status-{{.Status}}">{{.Status}}</td>
      <td class="dur">{{.DurationMs}}ms</td>
      <td class="msg">{{.Message}}</td>
    </tr>
  {{end}}
  </tbody>
</table>
</body>
</html>
`))
