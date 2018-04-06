package verifier

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/improbable-eng/thanos/pkg/block"
	"github.com/improbable-eng/thanos/pkg/compact"
	"github.com/improbable-eng/thanos/pkg/objstore"
	"github.com/pkg/errors"
	"github.com/prometheus/tsdb"
)

const OverlappedBlocksIssueID = "overlapped_blocks"

// OverlappedBlocksIssue checks bucket for blocks with overlapped time ranges.
// No repair is available for this issue.
func OverlappedBlocksIssue(ctx context.Context, logger log.Logger, bkt objstore.Bucket, _ objstore.Bucket, repair bool) error {
	level.Info(logger).Log("msg", "started verifying issue", "with-repair", repair, "issue", OverlappedBlocksIssueID)

	metas := map[string][]tsdb.BlockMeta{}
	err := bkt.Iter(ctx, "", func(name string) error {
		id, ok := block.IsBlockDir(name)
		if !ok {
			return nil
		}

		m, err := block.DownloadMeta(ctx, bkt, id)
		if err != nil {
			return err
		}

		metas[compact.GroupKey(m)] = append(metas[compact.GroupKey(m)], m.BlockMeta)
		return nil
	})
	if err != nil {
		return errors.Wrap(err, OverlappedBlocksIssueID)
	}

	overlaps := map[string]tsdb.Overlaps{}
	for k, groupMetas := range metas {
		o := tsdb.OverlappingBlocks(groupMetas)
		if len(o) > 0 {
			overlaps[k] = o
		}
	}

	if len(overlaps) == 0 {
		// All good.
		return nil
	}

	for k, o := range overlaps {
		level.Warn(logger).Log("msg", "found overlapped blocks", "group", k, "overlap", o)
	}

	if repair {
		level.Warn(logger).Log("msg", "repair is not implemented for this issue", "issue", OverlappedBlocksIssueID)
	}
	return nil
}