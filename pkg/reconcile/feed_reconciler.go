package reconcile

// FindMissingFeedEvents 对比现有收件箱与博主最新发布，计算缺失事件列表
func FindMissingFeedEvents(existingInbox, authorLatest []int64) []int64 {
	existingSet := make(map[int64]struct{}, len(existingInbox))
	for _, id := range existingInbox {
		existingSet[id] = struct{}{}
	}

	missing := make([]int64, 0)
	for _, id := range authorLatest {
		if _, exists := existingSet[id]; !exists {
			missing = append(missing, id)
		}
	}
	return missing
}
