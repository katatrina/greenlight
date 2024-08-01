package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/katatrina/greenlight/internal/db"
	"github.com/katatrina/greenlight/internal/util"
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
	Title   *string     `json:"title"`
	Year    *int32      `json:"year"`
	Runtime *db.Runtime `json:"runtime"`
	Genres  []string    `json:"genres"`
}

func validateUpdateMovieRequest(req *updateMovieRequest) validator.Violations {
	violations := validator.New()

	if req.Title != nil {
		if err := validator.ValidateMovieTitle(*req.Title); err != nil {
			violations.AddError("title", err.Error())
		}
	}

	if req.Year != nil {
		if err := validator.ValidateMovieYear(*req.Year); err != nil {
			violations.AddError("year", err.Error())
		}
	}

	if req.Runtime != nil {
		if err := validator.ValidateMovieRuntime(int32(*req.Runtime)); err != nil {
			violations.AddError("runtime", err.Error())
		}
	}

	if req.Genres != nil {
		if err := validator.ValidateMovieGenres(req.Genres); err != nil {
			violations.AddError("genres", err.Error())
		}
	}

	return violations
}

func (app *application) updateMovieHandler(ctx *gin.Context) {
	movieID, err := app.readIDParam(ctx)
	if err != nil {
		app.notFoundResponse(ctx)
		return
	}

	_, err = app.store.GetMovie(ctx, movieID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			app.notFoundResponse(ctx)
			return
		}

		app.serverErrorResponse(ctx, err)
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

	arg := db.UpdateMovieParams{
		Title: pgtype.Text{
			String: util.GetNullableString(req.Title),
			Valid:  req.Title != nil,
		},
		Year: pgtype.Int4{
			Int32: util.GetNullableInt32(req.Year),
			Valid: req.Year != nil,
		},
		Runtime: pgtype.Int4{
			Int32: int32(util.GetNullableRuntime(req.Runtime)),
			Valid: req.Runtime != nil,
		},
		Genres: req.Genres,
		ID:     movieID,
	}

	movie, err := app.store.UpdateMovie(ctx, arg)
	if err != nil {
		app.serverErrorResponse(ctx, err)
		return
	}

	rsp := envelop{"movie": movie}
	app.writeJSON(ctx, http.StatusOK, rsp, nil)
}

func (app *application) deleteMovieHandler(ctx *gin.Context) {
	movieID, err := app.readIDParam(ctx)
	if err != nil {
		app.notFoundResponse(ctx)
		return
	}

	rowsAffected, err := app.store.DeleteMovie(ctx, movieID)
	if err != nil {
		app.serverErrorResponse(ctx, err)
		return
	}

	// If no rows were affected, then the movie with that ID did not exist.
	if rowsAffected == 0 || rowsAffected != 1 {
		app.notFoundResponse(ctx)
		return
	}

	rsp := envelop{"message": "movie successfully deleted!"}
	app.writeJSON(ctx, http.StatusOK, rsp, nil)
}
