package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"suining-care/internal/response"
)

// POST /api/v1/vitals/latest —— 按姓名+时间范围分页查询。
func (h *Handler) listLatestVitals(w http.ResponseWriter, r *http.Request) {
	m, ok := bodyMap(w, r)
	if !ok {
		return
	}
	page, pageSize, ok := strictPage(w, m)
	if !ok {
		return
	}
	name := str(m, "name")
	timeFirst := str(m, "timeFirst")
	timeLast := str(m, "timeLast")
	sql := "SELECT * FROM (" + latestVitalsQuery + ") t WHERE 1=1"
	var args []any
	if strings.TrimSpace(name) != "" {
		sql += " AND t.name LIKE $" + strconv.Itoa(len(args)+1)
		args = append(args, "%"+name+"%")
	}
	if strings.TrimSpace(timeFirst) != "" && strings.TrimSpace(timeLast) != "" {
		sql += " AND t.createTime BETWEEN $" + strconv.Itoa(len(args)+1) + " AND $" + strconv.Itoa(len(args)+2)
		args = append(args, timeFirst, timeLast)
	}
	sql += " ORDER BY t.createTime DESC, t.oldId"
	total, ok := h.pagedTotal(w, r, sql, args)
	if !ok {
		return
	}
	limitSQL, fullArgs := pageArgs(args, page, pageSize)
	rows, err := h.DB.Query(r.Context(), sql+limitSQL, fullArgs...)
	if internalError(w, err) {
		return
	}
	defer rows.Close()
	items := []vitalRecord{}
	for rows.Next() {
		d, err := scanVitalRecord(rows)
		if err != nil {
			rows.Close()
			response.Error500(w, ref())
			return
		}
		items = append(items, d)
	}
	if !rowsOK(w, rows) {
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{
		"vitalRecords": items, "total": total, "page": page, "pageSize": pageSize,
	})
}

// POST /api/v1/vitals/history —— 家属查询自己老人的心率血压列表。
func (h *Handler) listVitalHistory(w http.ResponseWriter, r *http.Request) {
	m, ok := bodyMap(w, r)
	if !ok {
		return
	}
	oldID, ok := h.resolveOldID(w, r, m)
	if !ok {
		return
	}
	sql := `SELECT o.name AS name, o.old_id AS oldId, o.sex AS sex,` +
		` r.rate_state AS rateState, r.pre_state AS preState, r.hr AS hr, r.bp AS bp, r.id AS recordId,` +
		` CAST(r.rate_handled AS varchar) AS rateHandled, CAST(r.pre_handled AS varchar) AS preHandled,` +
		` TO_CHAR(r.create_time,'YYYY-MM-DD HH24:MI:SS') AS createTime` +
		` FROM oldperson o` +
		` INNER JOIN (SELECT * FROM ratepre WHERE old_id=$1 ORDER BY create_time DESC, id DESC LIMIT 1) r ON r.old_id=o.old_id` +
		` ORDER BY createTime DESC`
	rows, err := h.DB.Query(r.Context(), sql, oldID)
	if internalError(w, err) {
		return
	}
	defer rows.Close()
	items := []vitalRecord{}
	for rows.Next() {
		d, err := scanVitalRecord(rows)
		if err != nil {
			rows.Close()
			response.Error500(w, ref())
			return
		}
		items = append(items, d)
	}
	if !rowsOK(w, rows) {
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"vitalRecords": items})
}

// POST /api/v1/vitals/chart —— 家属查询老人心率血压明细（图表用）。
func (h *Handler) getVitalChart(w http.ResponseWriter, r *http.Request) {
	m, ok := bodyMap(w, r)
	if !ok {
		return
	}
	oldID, ok := h.resolveOldID(w, r, m)
	if !ok {
		return
	}
	rows, err := h.DB.Query(r.Context(), `SELECT r.hr AS hr, r.bp AS bp, TO_CHAR(r.create_time,'YYYY-MM-DD HH24:MI:SS') AS createTime
		FROM (SELECT id, hr, bp, create_time FROM ratepre WHERE old_id=$1 ORDER BY create_time DESC, id DESC LIMIT 30) r
		ORDER BY r.create_time ASC, r.id ASC`, oldID)
	if internalError(w, err) {
		return
	}
	defer rows.Close()
	type detail struct {
		Hr         *int64
		Bp         *int64
		CreateTime *string
	}
	var list []detail
	for rows.Next() {
		var d detail
		if err := rows.Scan(&d.Hr, &d.Bp, &d.CreateTime); err != nil {
			rows.Close()
			response.Error500(w, ref())
			return
		}
		list = append(list, d)
	}
	if !rowsOK(w, rows) {
		return
	}
	timeStr := ""
	lastDay := ""
	hrs := []*int64{}
	bps := []*int64{}
	createTimes := []string{}
	for _, d := range list {
		ct := ""
		if d.CreateTime != nil {
			ct = *d.CreateTime
		}
		if timeStr == "" {
			timeStr = safeSub(ct, 0, 10)
		}
		if ct != "" {
			lastDay = safeSub(ct, 0, 10)
		}
		hrs = append(hrs, d.Hr)
		bps = append(bps, d.Bp)
		createTimes = append(createTimes, ct)
	}
	if lastDay != "" && lastDay != timeStr {
		timeStr += " 至 " + lastDay
	}
	response.JSON(w, http.StatusOK, map[string]any{
		"time":        timeStr,
		"hrs":         hrs,
		"bps":         bps,
		"createTimes": createTimes,
	})
}
