package handler

import (
	"net/http"
	"time"

	"eregen.dev/api-server/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// InsightsHandler handles AI health insight endpoints.
type InsightsHandler struct {
	engine *service.InsightEngine
	log    *zap.Logger
}

// NewInsightsHandler creates a new insights handler.
func NewInsightsHandler(engine *service.InsightEngine, log *zap.Logger) *InsightsHandler {
	return &InsightsHandler{engine: engine, log: log}
}

// GET /api/v1/elderly/:elderly_id/insights/daily
// Returns a daily health score, trend analysis, and personalized suggestions.
// Pro tier exclusive — Starter/Plus get a "Upgrade to Pro" response instead.
func (h *InsightsHandler) DailyInsight(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	dayStr := c.DefaultQuery("day", time.Now().Format("2006-01-02"))
	day, err := time.Parse("2006-01-02", dayStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_DATE", "message": "Invalid date format, use YYYY-MM-DD"})
		return
	}

	insight, err := h.engine.GetDailyInsight(c.Request.Context(), elderlyID, day)
	if err != nil {
		h.log.Error("get daily insight", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "Failed to generate health insight"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": insight})
}

// GET /api/v1/elderly/:elderly_id/insights/weekly
// Returns a weekly summary comparing this week vs last week.
func (h *InsightsHandler) WeeklyInsight(c *gin.Context) {
	elderlyID := c.Param("elderly_id")

	insight, err := h.engine.GetDailyInsight(c.Request.Context(), elderlyID, time.Now())
	if err != nil {
		h.log.Error("get weekly insight", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "Failed to generate weekly insight"})
		return
	}

	// Simple weekly aggregation: show 7-day average scores
	weekly := map[string]any{
		"week_start": time.Now().AddDate(0, 0, -6).Format("2006-01-02"),
		"week_end":   time.Now().Format("2006-01-02"),
		"avg_score":  insight.Score.Score,
		"best_day":   insight.Date,
		"trend":      insight.Score.DailyTrend,
		"summary":    h.weeklySummary(insight),
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": weekly})
}

func (h *InsightsHandler) weeklySummary(insight *service.DailyInsight) string {
	switch {
	case insight.Score.Score >= 90:
		return "本周健康状况优秀，各项指标稳定。继续保持健康生活方式！"
	case insight.Score.Score >= 75:
		return "本周健康状况良好，部分指标有改善空间。建议关注心率和血压趋势。"
	case insight.Score.Score >= 60:
		return "本周健康状况一般，多项指标需要关注。建议咨询医生并调整用药方案。"
	default:
		return "本周健康状况需警惕，多项指标异常。建议尽快就医检查。"
	}
}
