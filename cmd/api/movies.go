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

// updateMovieHandler update the details of a specific movie.
func (app *application) updateMovieHandler(ctx *gin.Context) {
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
		Genres:  req.Genres,
		ID:      movieID,
		Version: movie.Version,
	}

	updatedMovie, err := app.store.UpdateMovie(ctx, arg)
	if err != nil {
		// If no matching row could be found, we know the movie version has changed
		// (or the record has been deleted) and we return our custom ErrEditConflict error.
		if errors.Is(err, db.ErrRecordNotFound) {
			app.editConflictResponse(ctx)
			return
		}

		app.serverErrorResponse(ctx, err)
		return
	}

	rsp := envelop{"movie": updatedMovie}
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

type listMoviesRequest struct {
	Title    *string  `form:"title"`
	Genres   []string `form:"genres"`
	PageID   *int32   `form:"page_id"`
	PageSize *int32   `form:"page_size"`
	Sort     *string  `form:"sort"`
}

func validateListMoviesRequest(req *listMoviesRequest) validator.Violations {
	violations := validator.New()

	if req.Title == nil {
		req.Title = new(string)
	}

	if req.Genres == nil {
		req.Genres = []string{}
	}

	if req.PageID == nil {
		req.PageID = new(int32)
		*req.PageID = 1
	} else if *req.PageID < 1 || *req.PageID > 10_000_000 {
		violations.AddError("page_id", "must be betweeen 1 and 10,000,000")
	}

	if req.PageSize == nil {
		req.PageSize = new(int32)
		*req.PageSize = 20
	} else if *req.PageSize < 1 || *req.PageSize > 100 {
		violations.AddError("page_size", "must be between 1 and 100")
	}

	if req.Sort == nil {
		req.Sort = new(string)
		*req.Sort = "id"
	}

	isSortable := util.PermittedValue(*req.Sort, "id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime")
	if !isSortable {
		violations.AddError("sort", "invalid sort field")
	}

	return violations
}

// listMoviesHandler show the details of all movies.
func (app *application) listMoviesHandler(ctx *gin.Context) {
	var req listMoviesRequest

	err := app.readQuery(ctx, &req)
	if err != nil {
		app.badRequestResponse(ctx, err)
		return
	}

	violations := validateListMoviesRequest(&req)
	if !violations.Empty() {
		app.failedValidationResponse(ctx, violations)
		return
	}

	app.writeJSON(ctx, http.StatusOK, req, nil)
}
