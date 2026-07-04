package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackietana/ticket-platform/catalog-service/internal/dto"
)

const MAX_MEMORY_SIZE = 5 * 1 << 20

func (h *Handler) GetEventByID(c *gin.Context) {
	strID := c.Param("id")
	if len(strID) == 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "missing id parameter"})
		return
	}

	id, err := uuid.Parse(strID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id parameter"})
		return
	}

	event, err := h.service.GetEventByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, event)
}

func (h *Handler) ListEvents(c *gin.Context) {
	var input dto.ListEventsInput

	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid input query"})
		return
	}

	events, err := h.service.ListEvents(c.Request.Context(), &input.CategoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, events)
}

func (h *Handler) CreateEvent(c *gin.Context) {
	var eventInput dto.CreateEventInput

	if err := c.ShouldBind(&eventInput); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	fileHeader, err := c.FormFile("poster")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	defer file.Close()

	eventInput.File = file
	eventInput.FileSize = fileHeader.Size
	eventInput.ContentType = fileHeader.Header.Get("Content-Type")

	event, err := h.service.CreateEvent(c.Request.Context(), eventInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.CreateEventResponse{ID: event.ID.String(), Message: "event successfully created"})
}
