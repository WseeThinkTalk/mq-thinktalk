package qa

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"mq-thinktalk/internal/svc"

	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

type ConsumeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewConsumeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConsumeLogic {
	return &ConsumeLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ConsumeLogic) Consume(ctx context.Context, key, val string) error {
	l.Infof("[Consume] qa event key: %s", key)

	var msg qaCanalMsg
	if err := json.Unmarshal([]byte(val), &msg); err != nil {
		l.Errorf("[Consume] unmarshal err: %v val: %s", err, val)
		return err
	}

	if msg.Table == "question" {
		return l.handleQuestion(ctx, &msg)
	}
	if msg.Table == "answer" {
		return l.handleAnswer(ctx, &msg)
	}
	return nil
}

func (l *ConsumeLogic) handleQuestion(ctx context.Context, msg *qaCanalMsg) error {
	for _, row := range msg.Data {
		var id int64
		var title, content, tagIds string
		var authorId, answerNum int64
		var status int
		var createTime string
		_ = json.Unmarshal(row["id"], &id)
		_ = json.Unmarshal(row["title"], &title)
		_ = json.Unmarshal(row["content"], &content)
		_ = json.Unmarshal(row["author_id"], &authorId)
		_ = json.Unmarshal(row["status"], &status)
		_ = json.Unmarshal(row["answer_num"], &answerNum)
		_ = json.Unmarshal(row["tag_ids"], &tagIds)
		_ = json.Unmarshal(row["create_time"], &createTime)

		if msg.Type == "DELETE" || status == 1 {
			l.deleteFromEs(ctx, "question-index", fmt.Sprintf("%d", id))
			continue
		}

		doc := map[string]interface{}{
			"id":          id,
			"title":       title,
			"content":     content,
			"author_id":   authorId,
			"status":      status,
			"answer_num":  answerNum,
			"tag_ids":     tagIds,
			"create_time": createTime,
		}
		l.upsertToEs(ctx, "question-index", fmt.Sprintf("%d", id), doc)
	}
	return nil
}

func (l *ConsumeLogic) handleAnswer(ctx context.Context, msg *qaCanalMsg) error {
	for _, row := range msg.Data {
		var id, questionId, authorId int64
		var content string
		var isAccepted, status int
		var likeNum, replyNum int64
		var createTime string
		_ = json.Unmarshal(row["id"], &id)
		_ = json.Unmarshal(row["question_id"], &questionId)
		_ = json.Unmarshal(row["author_id"], &authorId)
		_ = json.Unmarshal(row["content"], &content)
		_ = json.Unmarshal(row["is_accepted"], &isAccepted)
		_ = json.Unmarshal(row["status"], &status)
		_ = json.Unmarshal(row["like_num"], &likeNum)
		_ = json.Unmarshal(row["reply_num"], &replyNum)
		_ = json.Unmarshal(row["create_time"], &createTime)

		if msg.Type == "DELETE" || status == 1 {
			l.deleteFromEs(ctx, "answer-index", fmt.Sprintf("%d", id))
			continue
		}

		doc := map[string]interface{}{
			"id":          id,
			"question_id": questionId,
			"author_id":   authorId,
			"content":     content,
			"is_accepted": isAccepted,
			"status":      status,
			"like_num":    likeNum,
			"reply_num":   replyNum,
			"create_time": createTime,
		}
		l.upsertToEs(ctx, "answer-index", fmt.Sprintf("%d", id), doc)
	}
	return nil
}

func (l *ConsumeLogic) upsertToEs(ctx context.Context, index, docID string, doc map[string]interface{}) {
	if l.svcCtx.Es == nil {
		return
	}
	body, _ := json.Marshal(doc)
	payload := fmt.Sprintf(`{"doc":%s,"doc_as_upsert":true}`, string(body))

	bi, _ := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client: l.svcCtx.Es.Client,
		Index:  index,
	})
	_ = bi.Add(ctx, esutil.BulkIndexerItem{
		Action:     "update",
		DocumentID: docID,
		Body:       strings.NewReader(payload),
	})
	_ = bi.Close(ctx)
}

func (l *ConsumeLogic) deleteFromEs(ctx context.Context, index, docID string) {
	if l.svcCtx.Es == nil {
		return
	}
	bi, _ := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client: l.svcCtx.Es.Client,
		Index:  index,
	})
	_ = bi.Add(ctx, esutil.BulkIndexerItem{
		Action:     "delete",
		DocumentID: docID,
	})
	_ = bi.Close(ctx)
}

type qaCanalMsg struct {
	Type  string                       `json:"type"`
	Table string                       `json:"table"`
	Data  []map[string]json.RawMessage `json:"data"`
}

func Consumers(ctx context.Context, svcCtx *svc.ServiceContext) []service.Service {
	return []service.Service{
		kq.MustNewQueue(svcCtx.Config.QaKq, NewConsumeLogic(ctx, svcCtx)),
	}
}
