package metrics

import (
	"bufio"
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type metrics = struct {
	filesNew        *prometheus.HistogramVec
	filesChanged    *prometheus.HistogramVec
	filesUnmodified *prometheus.HistogramVec
	filesProcessed  *prometheus.HistogramVec

	dirsNew        *prometheus.HistogramVec
	dirsChanged    *prometheus.HistogramVec
	dirsUnmodified *prometheus.HistogramVec

	bytesAdded     *prometheus.HistogramVec // data_added
	bytesProcessed *prometheus.HistogramVec // total_bytes_processed

}

type JsonMetrics struct {
	MessageType         string  `json:"message_type"`
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

func ReadJson(jsonReader *bufio.Reader) (*JsonMetrics, error) {
	var stats JsonMetrics
	for {
		line, _, err := jsonReader.ReadLine()
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(line, &stats); err != nil {
			return nil, err
		}
		// just ignore progress json
		if stats.MessageType != "summary" {
			continue
		}
		return &stats, nil
	}
}

const ns, sub = "restic", "backup"

var (
	// TODO: allow this to be customized in the config
	labels       = []string{"repo""}
	sizeBuckets  = prometheus.ExponentialBuckets(256, 4, 8)
	filesChanged = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "backup_files_changed",
		Help:      "Total number of files changed.",
		Buckets:   sizeBuckets,
	}, labels)
	filesNew = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "backup_files_new",
		Help:      "Total number of files added.",
		Buckets:   sizeBuckets,
	}, labels)
	filesUnmodified = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "backup_files_unmodified",
		Help:      "Total number of files unmodified.",
		Buckets:   sizeBuckets,
	}, labels)
	filesProcessed = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "backup_files_processed",
		Help:      "Total number of files processed.",
		Buckets:   sizeBuckets,
	}, labels)
	dirsChanged = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "backup_dirs_changed",
		Help:      "Total number of dirs changed.",
		Buckets:   sizeBuckets,
	}, labels)
	dirsNew = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "backup_dirs_new",
		Help:      "Total number of dirs added.",
		Buckets:   sizeBuckets,
	}, labels)
	dirsUnmodified = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "backup_dirs_unmodified",
		Help:      "Total number of dirs unmodified.",
		Buckets:   sizeBuckets,
	}, labels)

	bytesAdded = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "backup_added_bytes",
		Help:      "Total number of bytes added.",
		Buckets:   sizeBuckets,
	}, labels)
	bytesProcessed = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "backup_processed_bytes",
		Help:      "Total number of bytes processed.",
		Buckets:   sizeBuckets,
	}, labels)
)
