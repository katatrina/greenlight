package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/greenlight/internal/db"
	"github.com/katatrina/greenlight/internal/validator"
)

type createMovieRequest struct {
	Title   string     `json:"title"`
	Year    int32      `json:"year"`
	Runtime db.Runtime `json:"runtime"`
	Genres  []string   `json:"genres"`
}

func validateCreateMovieRequest(req *createMovieRequest) validator.Violations {
	violations := validator.New()

	if err := validator.ValidateMovieTitle(req.Title); err != nil {
		violations.AddError("title", err.Error())
	}

	if err := validator.ValidateMovieYear(req.Year); err != nil {
		violations.AddError("year", err.Error())
	}

	if err := validator.ValidateMovieRuntime(int32(req.Runtime)); err != nil {
		violations.AddError("runtime", err.Error())
	}

	if err := validator.ValidateMovieGenres(req.Genres); err != nil {
		violations.AddError("genres", err.Error())
	}

	return violations
}

// createMovieHandler create a new movie.
func (app *application) createMovieHandler(ctx *gin.Context) {
	var req createMovieRequest

	err := app.readJSON(ctx, &req)
	if err != nil {
		app.badRequestResponse(ctx, err)
		return
	}

	violations := validateCreateMovieRequest(&req)
	if !violations.Empty() {
		app.failedValidationResponse(ctx, violations)
		return
	}

	movie, err := app.store.CreateMovie(ctx, db.CreateMovieParams{
		Title:   req.Title,
		Year:    req.Year,
		Runtime: req.Runtime,
		Genres:  req.Genres,
	})
	if err != nil {
		app.serverErrorResponse(ctx, err)
		return
	}

	// We want to include a Location header to let the client know
	// where they can find the newly-created movie resource at.
	headers := make(map[string]string)
	headers["Location"] = "/v1/movies/" + strconv.FormatInt(movie.ID, 10)

	rsp := envelop{"movie": movie}
	app.writeJSON(ctx, http.StatusCreated, rsp, headers)
}

// showMovieHandler show the details of a specific movie.
func (app *application) showMovieHandler(ctx *gin.Context) {
	// Try to convert the id string to a base 10 integer (with a bit size of 64).
	movieID, err := app.readIDParam(ctx)
	if err != nil {
		app.notFoundResponse(ctx)
		return
	}

	movie, err := app.store.GetMovie(ctx, movieID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			app.notFoundResponse(ctx)
			return
		}

		app.serverErrorResponse(ctx, err)
		return
	}

	rsp := envelop{"movie": movie}
	app.writeJSON(ctx, http.StatusOK, rsp, nil)
}

type updateMovieRequest struct {
	Title   string     `json:"title"`
	Year    int32      `json:"year"`
	Runtime db.Runtime `json:"runtime"`
	Genres  []string   `json:"genres"`
}

func validateUpdateMovieRequest(req *updateMovieRequest) validator.Violations {
	violations := validator.New()

	if err := validator.ValidateMovieTitle(req.Title); err != nil {
		violations.AddError("title", err.Error())
	}

	if err := validator.ValidateMovieYear(req.Year); err != nil {
		violations.AddError("year", err.Error())
	}

	if err := validator.ValidateMovieRuntime(int32(req.Runtime)); err != nil {
		violations.AddError("runtime", err.Error())
	}

	if err := validator.ValidateMovieGenres(req.Genres); err != nil {
		violations.AddError("genres", err.Error())
	}

	return violations
}

func (app *application) updateMovieHandler(ctx *gin.Context) {
	movieID, err := app.readIDParam(ctx)
	if err != nil {
		app.notFoundResponse(ctx)
		return
	}

	var req updateMovieRequest

	err = app.readJSON(ctx, &req)
	if err != nil {
		app.badRequestResponse(ctx, err)
		return
	}

	violations := validateUpdateMovieRequest(&req)
	if !violations.Empty() {
		app.failedValidationResponse(ctx, violations)
		return
	}

	movie, err := app.store.UpdateMovie(ctx, db.UpdateMovieParams{
		ID:      movieID,
		Title:   req.Title,
		Year:    req.Year,
		Runtime: req.Runtime,
		Genres:  req.Genres,
	})
	if err != nil {
		app.serverErrorResponse(ctx, err)
		return
	}

	rsp := envelop{"movie": movie}
	app.writeJSON(ctx, http.StatusOK, rsp, nil)
}
