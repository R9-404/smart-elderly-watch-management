package handlers

import (
	"context"
	"net/http"
	"strconv"
	"suining-care/internal/db"
	"suining-care/internal/notify"
	"suining-care/internal/response"
	"time"
)

// POST /api/v1/notices/create —— 发布公告（事务内执行；提交后异步发短信）。
func (h *Handler) createNotice(w http.ResponseWriter, r *http.Request) {
	m, ok := bodyMap(w, r)
	if !ok {
		return
	}
	title := str(m, "title")
	textMain := str(m, "textMain")
	fileList, ok := picNameList(w, m)
	if !ok {
		return
	}
	tx, err := h.DB.Begin(r.Context())
	if internalError(w, err) {
		return
	}
	defer tx.Rollback(r.Context())
	if !h.lockImages(w, r.Context(), tx, fileList) {
		return
	}
	picID, err := allocatePictureGroupID(r.Context(), tx)
	if internalError(w, err) {
		return
	}
	if !insertNoticePictures(w, r.Context(), tx, picID, fileList) {
		return
	}
	var textID int64
	err = tx.QueryRow(r.Context(), "INSERT INTO "+socialTextTable(tx.IsMySQL())+"(title,text_main,time,pic_id) VALUES($1,$2,$3,$4) RETURNING text_id",
		title, textMain, time.Now(), picID).Scan(&textID)
	if internalError(w, err) {
		return
	}
	phones, err := notificationPhones(r.Context(), tx)
	if internalError(w, err) {
		return
	}
	messages := make([]notify.Message, 0, len(phones))
	for _, target := range phones {
		messages = append(messages, notify.Message{Topic: "announcement", Key: strconv.FormatInt(textID, 10), FamilyID: target.FamilyID, Phone: target.Phone})
	}
	if internalError(w, notify.EnqueueBatch(r.Context(), tx, h.Box, messages)) {
		return
	}
	if internalError(w, tx.Commit(r.Context())) {
		return
	}
	// The announcement and durable outbox commit together; network I/O is worker-owned.
	writeCode(w, true)
}

// POST /api/v1/notices/delete —— 删除公告及关联图片（事务内执行）。
func (h *Handler) deleteNotice(w http.ResponseWriter, r *http.Request) {
	m, ok := bodyMap(w, r)
	if !ok {
		return
	}
	textID, ok := numParam(m, "textId")
	if !ok {
		response.Fail(w, "参数不正确")
		return
	}
	pictures, ok := h.noticePictureGroup(w, r, textID)
	if !ok {
		return
	}
	tx, err := h.DB.Begin(r.Context())
	if internalError(w, err) {
		return
	}
	defer tx.Rollback(r.Context())
	ct, err := tx.Exec(r.Context(), "DELETE FROM "+socialTextTable(tx.IsMySQL())+" WHERE text_id=$1", textID)
	if internalError(w, err) {
		return
	}
	if pictures != nil {
		if _, err := tx.Exec(r.Context(), "DELETE FROM pic WHERE pic_id=$1", *pictures); internalError(w, err) {
			return
		}
	}
	if ct.RowsAffected() != 1 {
		writeCode(w, false)
		return
	}
	if internalError(w, tx.Commit(r.Context())) {
		return
	}
	writeCode(w, true)
}

// POST /api/v1/notices/update —— 更新公告并替换图片（事务内执行）。
func (h *Handler) updateNotice(w http.ResponseWriter, r *http.Request) {
	m, ok := bodyMap(w, r)
	if !ok {
		return
	}
	textID, ok := numParam(m, "textId")
	if !ok {
		response.Fail(w, "参数不正确")
		return
	}
	title := str(m, "title")
	textMain := str(m, "textMain")
	pictures, ok := h.noticePictureGroup(w, r, textID)
	if !ok {
		return
	}
	fileList, ok := picNameList(w, m)
	if !ok {
		return
	}
	tx, err := h.DB.Begin(r.Context())
	if internalError(w, err) {
		return
	}
	defer tx.Rollback(r.Context())
	picID := int64(0)
	if !h.lockImages(w, r.Context(), tx, fileList) {
		return
	}
	if pictures == nil {
		picID, err = allocatePictureGroupID(r.Context(), tx)
		if internalError(w, err) {
			return
		}
	} else {
		picID = *pictures
		if _, err := tx.Exec(r.Context(), "DELETE FROM pic WHERE pic_id=$1", picID); internalError(w, err) {
			return
		}
	}
	if !insertNoticePictures(w, r.Context(), tx, picID, fileList) {
		return
	}
	ct, err := tx.Exec(r.Context(), "UPDATE "+socialTextTable(tx.IsMySQL())+" SET title=$1, text_main=$2, time=$3, pic_id=$4 WHERE text_id=$5",
		title, textMain, time.Now(), picID, textID)
	if internalError(w, err) {
		return
	}
	if ct.RowsAffected() != 1 {
		writeCode(w, false)
		return
	}
	if internalError(w, tx.Commit(r.Context())) {
		return
	}
	writeCode(w, true)
}

func (h *Handler) noticePictureGroup(w http.ResponseWriter, r *http.Request, id int64) (*int64, bool) {
	var pictures *int64
	err := h.DB.QueryRow(r.Context(), "SELECT pic_id FROM "+socialTextTable(h.DB.IsMySQL())+" WHERE text_id=$1", id).Scan(&pictures)
	if isNotFound(err) {
		writeCode(w, false)
		return nil, false
	}
	return pictures, !internalError(w, err)
}

func insertNoticePictures(w http.ResponseWriter, ctx context.Context, tx *db.Transaction, id int64, names []string) bool {
	for _, name := range names {
		result, err := tx.Exec(ctx, "INSERT INTO pic(pic_id,pic_name) VALUES($1,$2)", id, name)
		if err != nil || result.RowsAffected() != 1 {
			response.Error500(w, ref())
			return false
		}
	}
	return true
}
