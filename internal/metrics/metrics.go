package metrics

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"io"
	"os"
	"strings"
)

const ns, sub = "restic", "backup"

// New returns an error with the supplied message.
// New also records the stack trace at the point it was called.
func New(repo string, textFile string, files []string) *Prom {
	prom := &Prom{textFile: textFile, labelValues: append(files, repo)}
	labels := []string{}
	for i := 0; i < len(files); i++ {
		labels = append(labels, fmt.Sprintf("dir_%d", i))
	}
	labels = append(labels, "repo")
	prom.filesChanged = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "backup_files_changed",
		Help:      "Total number of files changed.",
	}, labels)
	prom.filesNew = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "backup_files_new",
		Help:      "Total number of files added.",
	}, labels)
	prom.filesUnmodified = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "backup_files_unmodified",
		Help:      "Total number of files unmodified.",
	}, labels)
	prom.filesProcessed = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "backup_files_processed",
		Help:      "Total number of files processed.",
	}, labels)
	prom.dirsChanged = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "backup_dirs_changed",
		Help:      "Total number of dirs changed.",
	}, labels)
	prom.dirsNew = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "backup_dirs_new",
		Help:      "Total number of dirs added.",
	}, labels)
	prom.dirsUnmodified = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "backup_dirs_unmodified",
		Help:      "Total number of dirs unmodified.",
	}, labels)

	prom.bytesAdded = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "backup_added_bytes",
		Help:      "Total number of bytes added.",
	}, labels)
	prom.bytesProcessed = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "backup_processed_bytes",
		Help:      "Total number of bytes processed.",
	}, labels)
	prom.errors = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "error_count",
		Help:      "number of errors occured",
	}, labels)
	prom.duration = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "duration_seconds",
		Help:      "backup duration in seconds",
	}, labels)
	return prom
}

type Prom struct {
	labelValues []string
	textFile string
	// Reads ErrorMetrics json from in, ignores unparsable json and copy it to
	// stderr
	numberErrors float64
	// stdout successfuly parsed?
	parsed bool
	filesChanged    *prometheus.GaugeVec
	filesNew        *prometheus.GaugeVec
	filesUnmodified *prometheus.GaugeVec
	filesProcessed  *prometheus.GaugeVec
	dirsChanged     *prometheus.GaugeVec
	dirsNew         *prometheus.GaugeVec
	dirsUnmodified  *prometheus.GaugeVec
	bytesAdded      *prometheus.GaugeVec
	bytesProcessed  *prometheus.GaugeVec
	errors          *prometheus.GaugeVec
	duration        *prometheus.GaugeVec
}

type MetricsError struct {
	Op   string `json:"Op"`
	Path string `json:"Path"`
	Err  int    `json:"Err"`
}

type MetricsErrorMessage struct {
	MessageType string `json:"message_type"`
	// error
	Error  MetricsError `json:"error"`
	During string       `json:"during"`
	Item   string       `json:"item"`
}

type Metrics struct {
	MessageType         string  `json:"message_type"`
	FilesNew            int     `json:"files_new"`
	FilesChanged        int     `json:"files_changed"`
	FilesUnmodified     int     `json:"files_unmodified"`
	DirsNew             int     `json:"dirs_new"`
	DirsChanged         int     `json:"dirs_changed"`
	DirsUnmodified      int     `json:"dirs_unmodified"`
	DataBlobs           int     `json:"data_blobs"`
	TreeBlobs           int     `json:"tree_blobs"`
	DataAdded           int     `json:"data_added"`
	TotalFilesProcessed int     `json:"total_files_processed"`
	TotalBytesProcessed int     `json:"total_bytes_processed"`
	TotalDuration       float64 `json:"total_duration"`
	SnapshotId          string  `json:"snapshot_id"`
}

// Collect JSON error messages unil EOF
func (p *Prom) CollectStderr(in *bufio.Reader) {
	var stats MetricsErrorMessage
	for {
		prompt, _ := in.Peek(31) // FIXME hardcoded: Peek for password prompt
		if (strings.HasPrefix(string(prompt),"enter password for repository: ")) {
			fmt.Fprintln(os.Stderr, string(prompt))
		}

		line, _, err := in.ReadLine()

		if err == io.EOF {
			p.errors.WithLabelValues(p.labelValues...).Set(p.numberErrors)
			return
		}

		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		if err := json.Unmarshal(line, &stats); err != nil {
			fmt.Fprintln(os.Stderr, string(line))
			continue
		}
		p.numberErrors++
	}
}

// Reads Metrics json from in, ignores unparsable json and copy it to
// stdout
func (p *Prom) CollectStdout(in *bufio.Reader) {
	var stats Metrics
	for {
		line, err := in.ReadBytes('\n')
		if err == io.EOF {
			return
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		if err := json.Unmarshal(line, &stats); err != nil {
			fmt.Fprint(os.Stdout, string(line))
			continue
		}
		if stats.MessageType != "summary" {
			continue
		}
		p.duration.WithLabelValues(p.labelValues...).Set(float64(stats.TotalDuration))
		p.filesNew.WithLabelValues(p.labelValues...).Set(float64(stats.FilesNew))
		p.filesUnmodified.WithLabelValues(p.labelValues...).Set(float64(stats.FilesUnmodified))
		p.filesChanged.WithLabelValues(p.labelValues...).Set(float64(stats.FilesChanged))
		p.dirsNew.WithLabelValues(p.labelValues...).Set(float64(stats.DirsNew))
		p.dirsChanged.WithLabelValues(p.labelValues...).Set(float64(stats.DirsChanged))
		p.dirsUnmodified.WithLabelValues(p.labelValues...).Set(float64(stats.DirsUnmodified))
		p.bytesAdded.WithLabelValues(p.labelValues...).Set(float64(stats.DataAdded))
		p.bytesProcessed.WithLabelValues(p.labelValues...).Set(float64(stats.TotalBytesProcessed))
		p.parsed = true
	}
}

func (p *Prom) WriteToTextFile() bool {
	if p.parsed {
		if err := prometheus.WriteToTextfile(p.textFile, prometheus.DefaultGatherer); err!=nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
	return p.parsed
}
