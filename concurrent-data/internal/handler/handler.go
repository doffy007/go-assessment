package handler

import (
	"net/http"

	"concurrent-data/internal/config"
	"concurrent-data/internal/service"

	"github.com/gin-gonic/gin"
)

type ProcessorHandler struct{}

// GetProgress godoc
// @Summary Get processing progress
// @Description Returns how many records have been processed so far
// @Tags Processor
// @Produce  json
// @Success 200 {object} map[string]interface{}
// @Router /progress [get]
func (h *ProcessorHandler) GetProgress(ctx *gin.Context) {
	progress := service.GetProgress()
	ctx.JSON(http.StatusOK, gin.H{
		"total":     progress.Total,
		"processed": progress.Processed,
		"percent":   progress.Percent,
	})
}

// ProcessFiles godoc
// @Summary Start concurrent CSV processing
// @Description Trigger worker pool to process all CSV files
// @Tags Processor
// @Produce json
// @Success 200 {array} service.Result
// @Router /process [post]
func (h *ProcessorHandler) ProcessFiles(ctx *gin.Context) {
	csvPath := config.AppConfig.CSVPath
	workers := config.AppConfig.Workers

	results, err := service.ProcessFiles(csvPath, workers)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "results": results})
		return
	}
	if len(results) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "no CSV files found"})
		return
	}
	ctx.JSON(http.StatusOK, results)
}
