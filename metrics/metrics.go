package metrics

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"os"
)

const ns, sub = "restic", "backup"

// New returns an error with the supplied message.
// New also records the stack trace at the point it was called.
func New(repo string, textFile string) *Prom {
	prom := &Prom{repo: repo}
	// TODO: allow this to be customized in the config
	labels := []string{"repo"}
	sizeBuckets := prometheus.ExponentialBuckets(256, 4, 8)
	prom.filesChanged = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "backup_files_changed",
		Help:      "Total number of files changed.",
		Buckets:   sizeBuckets,
	}, labels)
	prom.filesNew = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "backup_files_new",
		Help:      "Total number of files added.",
		Buckets:   sizeBuckets,
	}, labels)
	prom.filesUnmodified = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "backup_files_unmodified",
		Help:      "Total number of files unmodified.",
		Buckets:   sizeBuckets,
	}, labels)
	prom.filesProcessed = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "backup_files_processed",
		Help:      "Total number of files processed.",
		Buckets:   sizeBuckets,
	}, labels)
	prom.dirsChanged = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "backup_dirs_changed",
		Help:      "Total number of dirs changed.",
		Buckets:   sizeBuckets,
	}, labels)
	prom.dirsNew = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "backup_dirs_new",
		Help:      "Total number of dirs added.",
		Buckets:   sizeBuckets,
	}, labels)
	prom.dirsUnmodified = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "backup_dirs_unmodified",
		Help:      "Total number of dirs unmodified.",
		Buckets:   sizeBuckets,
	}, labels)

	prom.bytesAdded = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "backup_added_bytes",
		Help:      "Total number of bytes added.",
		Buckets:   sizeBuckets,
	}, labels)
	prom.bytesProcessed = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "backup_processed_bytes",
		Help:      "Total number of bytes processed.",
		Buckets:   sizeBuckets,
	}, labels)
	prom.errors = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "errors_total",
		Help:      "Total number of errors occured.",
		Buckets:   sizeBuckets,
	}, labels)
	return prom
}

type Prom struct {
	repo    string
	textFile string
	metrics promMetrics
	// Reads ErrorMetrics json from in, ignores unparsable json and copy it to
	// stderr
	numberErrors float64
	filesChanged *prometheus.HistogramVec
	filesNew     *prometheus.HistogramVec
	filesUnmodified *prometheus.HistogramVec
	filesProcessed  *prometheus.HistogramVec
	dirsChanged *prometheus.HistogramVec
	dirsNew *prometheus.HistogramVec
	dirsUnmodified *prometheus.HistogramVec
	bytesAdded *prometheus.HistogramVec
	bytesProcessed *prometheus.HistogramVec
	errors *prometheus.HistogramVec
}

type promMetrics = struct {
	filesNew        *prometheus.HistogramVec
	filesChanged    *prometheus.HistogramVec
	filesUnmodified *prometheus.HistogramVec
	filesProcessed  *prometheus.HistogramVec

	dirsNew        *prometheus.HistogramVec
	dirsChanged    *prometheus.HistogramVec
	dirsUnmodified *prometheus.HistogramVec

	bytesAdded     *prometheus.HistogramVec // data_added
	bytesProcessed *prometheus.HistogramVec // total_bytes_processed

	errors *prometheus.HistogramVec
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
	MessageType string `json:"message_type"`
	// progress
	FilesDone int `json:"files_done"`
	BytesDone int `json:"bytes_done"`
	// summary
	FilesNew            int     `json:"files_new"`
	FilesChaned         int     `json:"files_changed"`
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

func (p *Prom) ReadErrorMessage(in *bufio.Reader) (*MetricsErrorMessage, error) {
	var stats MetricsErrorMessage
	for {
		line, _, err := in.ReadLine()
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(line, &stats); err != nil {
			fmt.Fprintln(os.Stderr, string(line))
			continue
		}
		p.numberErrors++
		return &stats, nil
	}
}

// Reads Metrics json from in, ignores unparsable json and copy it to
// stdout
func (p Prom) ReadMessage(in *bufio.Reader) (*Metrics, error) {
	var stats Metrics
	for {
		line, _, err := in.ReadLine()
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(line, &stats); err != nil {
			fmt.Fprintln(os.Stdout, string(line))
			continue
		}
		return &stats, nil
	}
}

func (p *Prom) WriteToTextFile() {
	p.errors.WithLabelValues(p.repo).Observe(p.numberErrors)
	// FIXME: Atomic rename?
	prometheus.WriteToTextfile(p.textFile, prometheus.DefaultGatherer)
}
