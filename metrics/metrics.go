package metrics

import (
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

func New() *metrics {
	const ns, sub = "restic", "backup"
	labels := []string{"server", "repo", "snapshot_id"}
	// TODO: allow this to be customized in the config
	sizeBuckets := prometheus.ExponentialBuckets(256, 4, 8)
	m := metrics{
		filesChanged: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: ns,
			Subsystem: sub,
			Name:      "backup_files_changed",
			Help:      "Total number of files changed.",
			Buckets:   sizeBuckets,
		}, labels),
		filesNew: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: ns,
			Subsystem: sub,
			Name:      "backup_files_new",
			Help:      "Total number of files added.",
			Buckets:   sizeBuckets,
		}, labels),
		filesUnmodified: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: ns,
			Subsystem: sub,
			Name:      "backup_files_unmodified",
			Help:      "Total number of files unmodified.",
			Buckets:   sizeBuckets,
		}, labels),
		filesProcessed: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: ns,
			Subsystem: sub,
			Name:      "backup_files_processed",
			Help:      "Total number of files processed.",
			Buckets:   sizeBuckets,
		}, labels),


		dirsChanged: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: ns,
			Subsystem: sub,
			Name:      "backup_dirs_changed",
			Help:      "Total number of dirs changed.",
			Buckets:   sizeBuckets,
		}, labels),
		dirsNew: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: ns,
			Subsystem: sub,
			Name:      "backup_dirs_new",
			Help:      "Total number of dirs added.",
			Buckets:   sizeBuckets,
		}, labels),
		dirsUnmodified: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: ns,
			Subsystem: sub,
			Name:      "backup_dirs_unmodified",
			Help:      "Total number of dirs unmodified.",
			Buckets:   sizeBuckets,
		}, labels),

		bytesAdded: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: ns,
			Subsystem: sub,
			Name:      "backup_added_bytes",
			Help:      "Total number of bytes added.",
			Buckets:   sizeBuckets,
		}, labels),
		bytesProcessed: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: ns,
			Subsystem: sub,
			Name:      "backup_processed_bytes",
			Help:      "Total number of bytes processed.",
			Buckets:   sizeBuckets,
		}, labels),
	}
	return &m
}
