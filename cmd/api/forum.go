// Filename: cmd/api/forum.go

package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"AWD_FinalProject.ryanarmstrong.net/internal/data"
	"AWD_FinalProject.ryanarmstrong.net/internal/validator"
)

// createForumHandler for the "Post /v1/forums" endpoint
func (app *application) createForumHandler(w http.ResponseWriter, r *http.Request) {
	// Our target decode destination
	var input struct {
		Topic      string `json:"topic"`
		Discussion string `json:"discussion"`
	}
	// Initialize a new json.Decoder instance
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Copy the values from the input struct to a new Forum struct
	forum := &data.Forum{
		Topic:      input.Topic,
		Discussion: input.Discussion,
	}
	// Initialize a new Validator instance
	v := validator.New()

	// Check the map to determine if there were any validation errors
	if data.ValidateForum(v, forum); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Create a Forum
	err = app.models.Forums.Insert(forum)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	// Create a Location header for the newly created resource/Forum
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/forums/%d", forum.ID))
	// Write the JSON response with 201 - Created status code with the body
	// being the Forum data and the header being the headers map
	err = app.writeJSON(w, http.StatusCreated, envelope{"forum": forum}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// showForumHandler for the "Post /v1/forums/:id" endpoint
func (app *application) showForumHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Fetch the specific forum
	forum, err := app.models.Forums.Get(id)
	// Handle errors
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Write the data returned by Get()
	err = app.writeJSON(w, http.StatusOK, envelope{"forum": forum}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateForumHandler(w http.ResponseWriter, r *http.Request) {
	// This method does a partial replacement
	// Get the id for the forum that needs updating
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Fetch the original record from the database
	forum, err := app.models.Forums.Get(id)
	// Handle errors
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Create an input struct to hold data read in from the client
	// We update input struct to use pointers because pointers have a
	// default value of nil
	// If a field remains nil then we know the client did not update it
	var input struct {
		Topic      *string `json:"topic"`
		Discussion *string `json:"discussion"`
	}
	// Initialize a new json.Decoder instance
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// Check for updates
	if input.Topic != nil {
		forum.Topic = *input.Topic
	}
	if input.Discussion != nil {
		forum.Discussion = *input.Discussion
		//forum.Mode = append(forum.Mode, strings.Join(input.Mode, " "))
	}
	// Perform validation on the updated Forum. If validation fails, then
	// we send a 422 - Unprocessable Entity response to the client
	// Initialize a new Validator instance
	v := validator.New()

	// Check the map to determine if there were any validation errors
	if data.ValidateForum(v, forum); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Pass the updated Forum record to the Update() method
	err = app.models.Forums.Update(forum)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Write the data returned by Update()
	err = app.writeJSON(w, http.StatusOK, envelope{"forum": forum}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteForumHandler(w http.ResponseWriter, r *http.Request) {
	// Get the id for the forum that needs to be deleted
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Delete the Forum from the database. Send a 404 Not Found status code to the
	// client if there is no matching record
	err = app.models.Forums.Delete(id)
	// Handle errors
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Return 200 Status OK to the client with a successful message
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "forum successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// The listForumsHandler allows the client to see a listing of forums
// based on a set of criteria
func (app *application) listForumsHandler(w http.ResponseWriter, r *http.Request) {
	// Create an input struct to hold our query parameters
	var input struct {
		Topic    string
		Comments []string
		data.Filters
	}
	// Initialize a validator
	v := validator.New()
	// Get the URL values map
	qs := r.URL.Query()
	// Use the helper methods to extract the values
	input.Topic = app.readString(qs, "topic", "")
	input.Comments = app.readCSV(qs, "comments", []string{})
	// Get the page information
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	// Get the sort information
	input.Filters.Sort = app.readString(qs, "sort", "id")
	// Specify the allowed sort values
	input.Filters.SortList = []string{"id", "topic", "created_at", "total_comments", "-id", "-topic", "-created_at", "-total_comments"}
	// Check for validation errors
	if data.ValidateFilers(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Get a listing of all forums
	forums, metadata, err := app.models.Forums.GetAll(input.Topic, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Send a JSON response containing all the forums
	err = app.writeJSON(w, http.StatusOK, envelope{"forums": forums, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) addCommentHandler(w http.ResponseWriter, r *http.Request) {
	// This method adds a comment to the slice
	// Get the id for the forum that the comment will be added
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Fetch the original record from the database
	forum, err := app.models.Forums.Get(id)
	// Handle errors
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Create an input struct to hold data read in from the client
	// We update input struct to use pointers because pointers have a
	// default value of nil
	// If a field remains nil then we know the client did not add a comment
	var input struct {
		Comments []string `json:"comments"`
	}
	// Initialize a new json.Decoder instance
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// Check for added discussion
	if input.Comments != nil {
		//forum.Comments = *input.Comments
		forum.Comments = append(forum.Comments, strings.Join(input.Comments, " "))
	}
	// Perform validation on the updated Forum. If validation fails, then
	// we send a 422 - Unprocessable Entity response to the client
	// Initialize a new Validator instance
	v := validator.New()

	// Check the map to determine if there were any validation errors
	if data.ValidateForum(v, forum); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Pass the updated Forum record to the addComment() method
	err = app.models.Forums.AddComment(forum)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Write the data returned by Update()
	err = app.writeJSON(w, http.StatusOK, envelope{"forum": forum}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
