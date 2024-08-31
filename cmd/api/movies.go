package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/katatrina/greenlight/internal/db"
	"github.com/katatrina/greenlight/internal/util"
	"github.com/katatrina/greenlight/internal/validator"
)

type createMovieRequest struct {
	Title       string     `json:"title"`
	PublishYear int32      `json:"publish_year"`
	Runtime     db.Runtime `json:"runtime"`
	Genres      []string   `json:"genres"`
}

func validateCreateMovieRequest(req *createMovieRequest) validator.Violations {
	violations := validator.New()

	if err := validator.ValidateMovieTitle(req.Title); err != nil {
		violations.AddError("title", err.Error())
	}

	if err := validator.ValidateMovieYear(req.PublishYear); err != nil {
		violations.AddError("publish_year", err.Error())
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

	// Parse the request body.
	err := app.readJSON(ctx, &req)
	if err != nil {
		app.badRequestResponse(ctx, err)
		return
	}

	// Validate the request body.
	violations := validateCreateMovieRequest(&req)
	if !violations.Empty() {
		app.failedValidationResponse(ctx, violations)
		return
	}

	// Create a new movie record in the database.
	movie, err := app.store.CreateMovie(ctx, db.CreateMovieParams{
		Title:       req.Title,
		PublishYear: req.PublishYear,
		Runtime:     req.Runtime,
		Genres:      req.Genres,
	})
	if err != nil {
		app.serverErrorResponse(ctx, err)
		return
	}

	// We want to include a Location header to let the client know
	// where they can find the newly-created movie resource at.
	headers := make(map[string]string)
	headers["Location"] = "/v1/movies/" + strconv.FormatInt(movie.ID, 10)

	rsp := envelope{"movie": movie}
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

	// Retrieve the movie record based on the provided ID.
	movie, err := app.store.GetMovie(ctx, movieID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			app.notFoundResponse(ctx)
			return
		}

		app.serverErrorResponse(ctx, err)
		return
	}

	rsp := envelope{"movie": movie}
	app.writeJSON(ctx, http.StatusOK, rsp, nil)
}

type updateMovieRequest struct {
	Title       *string     `json:"title"`
	PublishYear *int32      `json:"publish_year"`
	Runtime     *db.Runtime `json:"runtime"`
	Genres      []string    `json:"genres"`
}

func validateUpdateMovieRequest(req *updateMovieRequest) validator.Violations {
	violations := validator.New()

	if req.Title != nil {
		if err := validator.ValidateMovieTitle(*req.Title); err != nil {
			violations.AddError("title", err.Error())
		}
	}

	if req.PublishYear != nil {
		if err := validator.ValidateMovieYear(*req.PublishYear); err != nil {
			violations.AddError("publish_year", err.Error())
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
	// Read the movie ID from the URL parameter.
	movieID, err := app.readIDParam(ctx)
	if err != nil {
		app.notFoundResponse(ctx)
		return
	}

	// Retrieve the existing movie record from the database.
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

	// Parse the request body.
	err = app.readJSON(ctx, &req)
	if err != nil {
		app.badRequestResponse(ctx, err)
		return
	}

	// Validate the request body.
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
		PublishYear: pgtype.Int4{
			Int32: util.GetNullableInt32(req.PublishYear),
			Valid: req.PublishYear != nil,
		},
		Runtime: pgtype.Int4{
			Int32: int32(util.GetNullableRuntime(req.Runtime)),
			Valid: req.Runtime != nil,
		},
		Genres:  req.Genres,
		ID:      movieID,
		Version: movie.Version,
	}

	// Try to update the movie.
	updatedMovie, err := app.store.UpdateMovie(ctx, arg)
	if err != nil {
		// If no matching row could be found, we know the movie's version has changed
		// (or the record has been deleted) and we we invoke the editConflictResponse method.
		if errors.Is(err, db.ErrRecordNotFound) {
			app.editConflictResponse(ctx)
			return
		}

		app.serverErrorResponse(ctx, err)
		return
	}

	rsp := envelope{"updated_movie": updatedMovie}
	app.writeJSON(ctx, http.StatusOK, rsp, nil)
}

// deleteMovieHandler delete a specific movie.
func (app *application) deleteMovieHandler(ctx *gin.Context) {
	movieID, err := app.readIDParam(ctx)
	if err != nil {
		app.notFoundResponse(ctx)
		return
	}

	// Attempt to delete the movie.
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

	rsp := envelope{"message": "movie successfully deleted!"}
	app.writeJSON(ctx, http.StatusOK, rsp, nil)
}

type listMoviesRequest struct {
	Title    string   `form:"title"`
	Genres   []string `form:"genres"`
	Page     *int32   `form:"page"`
	PageSize *int32   `form:"page_size"`
	Sort     string   `form:"sort"`
}

type listMoviesResponse struct {
	Metadata db.PaginationMetadata `json:"metadata"`
	Movies   []db.Movie            `json:"movies"`
}

func newListMoviesWithFiltersRowResponse(row db.ListMoviesWithFiltersRow) db.Movie {
	return db.Movie{
		ID:          row.Movie.ID,
		Title:       row.Movie.Title,
		Runtime:     row.Movie.Runtime,
		Genres:      row.Movie.Genres,
		PublishYear: row.Movie.PublishYear,
		Version:     row.Movie.Version,
		CreatedAt:   row.Movie.CreatedAt,
	}
}

// validateListMoviesRequest validates the listMoviesRequest struct and sets default "fallback" values if necessary.
func validateListMoviesRequest(req *listMoviesRequest) validator.Violations {
	violations := validator.New()

	// If the genres field is not provided, set it to an empty slice.
	if req.Genres == nil {
		req.Genres = []string{}
	} else if req.Genres[0] == "" { // If the genres field is provided but empty, also set it to an empty slice.
		req.Genres = []string{}
	} else { // If the genres field is provided, split the comma-separated string into a slice.
		req.Genres = strings.Split(req.Genres[0], ",")
	}

	if req.Page == nil { // If the page_id is not provided, set it to 1.
		req.Page = new(int32)
		*req.Page = 1
	} else if !(*req.Page >= 1 && *req.Page <= 10_000_000) {
		violations.AddError("page", "must be betweeen 1 and 10,000,000")
	}

	if req.PageSize == nil { // If the page_size is not provided, set it to 20.
		req.PageSize = new(int32)
		*req.PageSize = 20
	} else if !(*req.PageSize >= 1 && *req.PageSize <= 100) {
		violations.AddError("page_size", "must be between 1 and 100")
	}

	// If the sort field is not provided, set it to "id".
	if req.Sort == "" {
		req.Sort = "id"
	}

	// Check if the sort field is one of the permitted values.
	sortSafeList := []string{"id", "title", "publishYear", "runtime", "-id", "-title", "-publishYear", "-runtime"}
	isSortable := util.PermittedValue(req.Sort, sortSafeList...)
	if !isSortable {
		violations.AddError("sort", fmt.Sprintf("invalid sort value <%s>", req.Sort))
	}

	return violations
}

// listMoviesHandler show the details of filtered movies.
func (app *application) listMoviesHandler(ctx *gin.Context) {
	var req listMoviesRequest

	// Parse query parameters
	err := app.readQueryParams(ctx, &req)
	if err != nil {
		app.badRequestResponse(ctx, err)
		return
	}

	// Validate query parameters
	violations := validateListMoviesRequest(&req)
	if !violations.Empty() {
		app.failedValidationResponse(ctx, violations)
		return
	}

	arg := db.ListMoviesWithFiltersParams{
		Title:   req.Title,
		Genres:  req.Genres,
		Reverse: strings.HasPrefix(req.Sort, "-"),
		OrderBy: strings.TrimPrefix(req.Sort, "-"),
		Limit:   *req.PageSize,
		Offset:  (*req.Page - 1) * *req.PageSize,
	}

	// Retrieve the list of movies based on the provided filters.
	movies, err := app.store.ListMoviesWithFilters(ctx, arg)
	if err != nil {
		app.serverErrorResponse(ctx, err)
		return
	}

	rsp := listMoviesResponse{
		Metadata: db.PaginationMetadata{},
		Movies:   make([]db.Movie, 0),
	}

	if len(movies) > 0 {
		rsp.Metadata = db.CalculatePaginationMetadata(movies[0].TotalRecords, *req.Page, *req.PageSize)
		// TODO: Think about a better solution to handle this.
		// Instead of using a for loop to just remove the "total_records" field from each movie.
		// We can use the "-" struct tag to exclude the field from the response.
		// However, currently, sqlc does not support custom struct tags.
		for _, v := range movies {
			rsp.Movies = append(rsp.Movies, newListMoviesWithFiltersRowResponse(v))
		}
	}

	/*
		We did scan the dataset into a slice and then loop over it just to exclude the "total_records" field.
		Double loop can cause performance issues when the dataset is large.
	*/

	app.writeJSON(ctx, http.StatusOK, rsp, nil)
}
