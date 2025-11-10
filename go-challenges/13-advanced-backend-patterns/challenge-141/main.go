package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/graphql-go/graphql"
)

// Domain Models
type Genre string

const (
	GenreFiction    Genre = "FICTION"
	GenreNonfiction Genre = "NONFICTION"
	GenreSciFi      Genre = "SCIFI"
	GenreFantasy    Genre = "FANTASY"
	GenreMystery    Genre = "MYSTERY"
	GenreRomance    Genre = "ROMANCE"
)

type Book struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	ISBN          string    `json:"isbn"`
	PublishedYear int       `json:"publishedYear"`
	AuthorID      string    `json:"authorId"`
	Genre         Genre     `json:"genre"`
	Rating        *float64  `json:"rating,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
}

type Author struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	Bio  *string `json:"bio,omitempty"`
}

type BookFilter struct {
	Genre     *Genre
	MinRating *float64
	Title     *string
	AuthorID  *string
}

type PaginationInput struct {
	First  *int
	After  *string
	Last   *int
	Before *string
}

type BookEdge struct {
	Node   *Book  `json:"node"`
	Cursor string `json:"cursor"`
}

type PageInfo struct {
	HasNextPage     bool    `json:"hasNextPage"`
	HasPreviousPage bool    `json:"hasPreviousPage"`
	StartCursor     *string `json:"startCursor,omitempty"`
	EndCursor       *string `json:"endCursor,omitempty"`
}

type BookConnection struct {
	Edges    []*BookEdge `json:"edges"`
	PageInfo *PageInfo   `json:"pageInfo"`
}

// Storage Layer
type Store struct {
	mu         sync.RWMutex
	books      map[string]*Book
	authors    map[string]*Author
	nextBookID int
	nextAuthorID int
}

func NewStore() *Store {
	return &Store{
		books:   make(map[string]*Book),
		authors: make(map[string]*Author),
		nextBookID: 1,
		nextAuthorID: 1,
	}
}

func (s *Store) CreateBook(title, isbn string, publishedYear int, authorID string, genre Genre, rating *float64) (*Book, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate author exists
	if _, exists := s.authors[authorID]; !exists {
		return nil, errors.New("author not found")
	}

	// Validate input
	if title == "" {
		return nil, errors.New("title is required")
	}
	if isbn == "" {
		return nil, errors.New("ISBN is required")
	}
	if publishedYear < 1000 || publishedYear > time.Now().Year()+1 {
		return nil, errors.New("invalid published year")
	}
	if rating != nil && (*rating < 0 || *rating > 5) {
		return nil, errors.New("rating must be between 0 and 5")
	}

	book := &Book{
		ID:            fmt.Sprintf("%d", s.nextBookID),
		Title:         title,
		ISBN:          isbn,
		PublishedYear: publishedYear,
		AuthorID:      authorID,
		Genre:         genre,
		Rating:        rating,
		CreatedAt:     time.Now(),
	}
	s.books[book.ID] = book
	s.nextBookID++
	return book, nil
}

func (s *Store) GetBook(id string) (*Book, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	book, exists := s.books[id]
	if !exists {
		return nil, errors.New("book not found")
	}
	return book, nil
}

func (s *Store) GetBooks(filter *BookFilter, pagination *PaginationInput) (*BookConnection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Collect all books
	var allBooks []*Book
	for _, book := range s.books {
		allBooks = append(allBooks, book)
	}

	// Apply filters
	var filtered []*Book
	for _, book := range allBooks {
		if filter != nil {
			if filter.Genre != nil && book.Genre != *filter.Genre {
				continue
			}
			if filter.MinRating != nil && (book.Rating == nil || *book.Rating < *filter.MinRating) {
				continue
			}
			if filter.Title != nil && !strings.Contains(strings.ToLower(book.Title), strings.ToLower(*filter.Title)) {
				continue
			}
			if filter.AuthorID != nil && book.AuthorID != *filter.AuthorID {
				continue
			}
		}
		filtered = append(filtered, book)
	}

	// Apply pagination
	return s.paginateBooks(filtered, pagination)
}

func (s *Store) paginateBooks(books []*Book, pagination *PaginationInput) (*BookConnection, error) {
	if len(books) == 0 {
		return &BookConnection{
			Edges:    []*BookEdge{},
			PageInfo: &PageInfo{HasNextPage: false, HasPreviousPage: false},
		}, nil
	}

	// Default pagination
	first := 10
	if pagination != nil && pagination.First != nil {
		first = *pagination.First
	}

	startIdx := 0
	if pagination != nil && pagination.After != nil {
		// Decode cursor
		afterID, err := decodeCursor(*pagination.After)
		if err == nil {
			for i, book := range books {
				if book.ID == afterID {
					startIdx = i + 1
					break
				}
			}
		}
	}

	endIdx := startIdx + first
	if endIdx > len(books) {
		endIdx = len(books)
	}

	// Create edges
	var edges []*BookEdge
	for i := startIdx; i < endIdx; i++ {
		edges = append(edges, &BookEdge{
			Node:   books[i],
			Cursor: encodeCursor(books[i].ID),
		})
	}

	// Page info
	pageInfo := &PageInfo{
		HasNextPage:     endIdx < len(books),
		HasPreviousPage: startIdx > 0,
	}

	if len(edges) > 0 {
		startCursor := edges[0].Cursor
		endCursor := edges[len(edges)-1].Cursor
		pageInfo.StartCursor = &startCursor
		pageInfo.EndCursor = &endCursor
	}

	return &BookConnection{
		Edges:    edges,
		PageInfo: pageInfo,
	}, nil
}

func (s *Store) UpdateBook(id, title, isbn string, publishedYear *int, rating *float64) (*Book, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	book, exists := s.books[id]
	if !exists {
		return nil, errors.New("book not found")
	}

	if title != "" {
		book.Title = title
	}
	if isbn != "" {
		book.ISBN = isbn
	}
	if publishedYear != nil {
		if *publishedYear < 1000 || *publishedYear > time.Now().Year()+1 {
			return nil, errors.New("invalid published year")
		}
		book.PublishedYear = *publishedYear
	}
	if rating != nil {
		if *rating < 0 || *rating > 5 {
			return nil, errors.New("rating must be between 0 and 5")
		}
		book.Rating = rating
	}

	return book, nil
}

func (s *Store) DeleteBook(id string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.books[id]; !exists {
		return false, errors.New("book not found")
	}

	delete(s.books, id)
	return true, nil
}

func (s *Store) CreateAuthor(name string, bio *string) (*Author, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if name == "" {
		return nil, errors.New("name is required")
	}

	author := &Author{
		ID:   fmt.Sprintf("%d", s.nextAuthorID),
		Name: name,
		Bio:  bio,
	}
	s.authors[author.ID] = author
	s.nextAuthorID++
	return author, nil
}

func (s *Store) GetAuthor(id string) (*Author, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	author, exists := s.authors[id]
	if !exists {
		return nil, errors.New("author not found")
	}
	return author, nil
}

func (s *Store) GetAuthors(limit *int) ([]*Author, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var authors []*Author
	for _, author := range s.authors {
		authors = append(authors, author)
	}

	if limit != nil && *limit > 0 && *limit < len(authors) {
		authors = authors[:*limit]
	}

	return authors, nil
}

func (s *Store) GetBooksByAuthor(authorID string) ([]*Book, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var books []*Book
	for _, book := range s.books {
		if book.AuthorID == authorID {
			books = append(books, book)
		}
	}
	return books, nil
}

// Cursor encoding/decoding
func encodeCursor(id string) string {
	return base64.StdEncoding.EncodeToString([]byte(id))
}

func decodeCursor(cursor string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GraphQL Schema
func buildSchema(store *Store) (graphql.Schema, error) {
	// Enum types
	genreEnum := graphql.NewEnum(graphql.EnumConfig{
		Name: "Genre",
		Values: graphql.EnumValueConfigMap{
			"FICTION":    &graphql.EnumValueConfig{Value: GenreFiction},
			"NONFICTION": &graphql.EnumValueConfig{Value: GenreNonfiction},
			"SCIFI":      &graphql.EnumValueConfig{Value: GenreSciFi},
			"FANTASY":    &graphql.EnumValueConfig{Value: GenreFantasy},
			"MYSTERY":    &graphql.EnumValueConfig{Value: GenreMystery},
			"ROMANCE":    &graphql.EnumValueConfig{Value: GenreRomance},
		},
	})

	// Object types (declared first for forward references)
	var bookType *graphql.Object
	var authorType *graphql.Object

	// Author type
	authorType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Author",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.NewNonNull(graphql.ID),
			},
			"name": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"bio": &graphql.Field{
				Type: graphql.String,
			},
			"books": &graphql.Field{
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.NewObject(graphql.ObjectConfig{
					Name: "BookRef",
					Fields: graphql.Fields{
						"id":            &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
						"title":         &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
						"isbn":          &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
						"publishedYear": &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
						"genre":         &graphql.Field{Type: graphql.NewNonNull(genreEnum)},
						"rating":        &graphql.Field{Type: graphql.Float},
						"createdAt":     &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
					},
				})))),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if author, ok := p.Source.(*Author); ok {
						return store.GetBooksByAuthor(author.ID)
					}
					return nil, errors.New("invalid source type")
				},
			},
		},
	})

	// Book type
	bookType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Book",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.NewNonNull(graphql.ID),
			},
			"title": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"isbn": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"publishedYear": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Int),
			},
			"genre": &graphql.Field{
				Type: graphql.NewNonNull(genreEnum),
			},
			"rating": &graphql.Field{
				Type: graphql.Float,
			},
			"createdAt": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if book, ok := p.Source.(*Book); ok {
						return book.CreatedAt.Format(time.RFC3339), nil
					}
					return nil, nil
				},
			},
			"author": &graphql.Field{
				Type: graphql.NewNonNull(authorType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if book, ok := p.Source.(*Book); ok {
						return store.GetAuthor(book.AuthorID)
					}
					return nil, errors.New("invalid source type")
				},
			},
		},
	})

	// PageInfo type
	pageInfoType := graphql.NewObject(graphql.ObjectConfig{
		Name: "PageInfo",
		Fields: graphql.Fields{
			"hasNextPage": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Boolean),
			},
			"hasPreviousPage": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Boolean),
			},
			"startCursor": &graphql.Field{
				Type: graphql.String,
			},
			"endCursor": &graphql.Field{
				Type: graphql.String,
			},
		},
	})

	// BookEdge type
	bookEdgeType := graphql.NewObject(graphql.ObjectConfig{
		Name: "BookEdge",
		Fields: graphql.Fields{
			"node": &graphql.Field{
				Type: graphql.NewNonNull(bookType),
			},
			"cursor": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
	})

	// BookConnection type
	bookConnectionType := graphql.NewObject(graphql.ObjectConfig{
		Name: "BookConnection",
		Fields: graphql.Fields{
			"edges": &graphql.Field{
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(bookEdgeType))),
			},
			"pageInfo": &graphql.Field{
				Type: graphql.NewNonNull(pageInfoType),
			},
		},
	})

	// Input types
	bookFilterInput := graphql.NewInputObject(graphql.InputObjectConfig{
		Name: "BookFilter",
		Fields: graphql.InputObjectConfigFieldMap{
			"genre":     &graphql.InputObjectFieldConfig{Type: genreEnum},
			"minRating": &graphql.InputObjectFieldConfig{Type: graphql.Float},
			"title":     &graphql.InputObjectFieldConfig{Type: graphql.String},
			"authorId":  &graphql.InputObjectFieldConfig{Type: graphql.String},
		},
	})

	paginationInput := graphql.NewInputObject(graphql.InputObjectConfig{
		Name: "PaginationInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"first":  &graphql.InputObjectFieldConfig{Type: graphql.Int},
			"after":  &graphql.InputObjectFieldConfig{Type: graphql.String},
			"last":   &graphql.InputObjectFieldConfig{Type: graphql.Int},
			"before": &graphql.InputObjectFieldConfig{Type: graphql.String},
		},
	})

	createBookInput := graphql.NewInputObject(graphql.InputObjectConfig{
		Name: "CreateBookInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"title":         &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
			"isbn":          &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
			"publishedYear": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.Int)},
			"authorId":      &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
			"genre":         &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(genreEnum)},
			"rating":        &graphql.InputObjectFieldConfig{Type: graphql.Float},
		},
	})

	updateBookInput := graphql.NewInputObject(graphql.InputObjectConfig{
		Name: "UpdateBookInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"title":         &graphql.InputObjectFieldConfig{Type: graphql.String},
			"isbn":          &graphql.InputObjectFieldConfig{Type: graphql.String},
			"publishedYear": &graphql.InputObjectFieldConfig{Type: graphql.Int},
			"rating":        &graphql.InputObjectFieldConfig{Type: graphql.Float},
		},
	})

	createAuthorInput := graphql.NewInputObject(graphql.InputObjectConfig{
		Name: "CreateAuthorInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"name": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
			"bio":  &graphql.InputObjectFieldConfig{Type: graphql.String},
		},
	})

	// Query type
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"books": &graphql.Field{
				Type: graphql.NewNonNull(bookConnectionType),
				Args: graphql.FieldConfigArgument{
					"filter":     &graphql.ArgumentConfig{Type: bookFilterInput},
					"pagination": &graphql.ArgumentConfig{Type: paginationInput},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var filter *BookFilter
					if f, ok := p.Args["filter"].(map[string]interface{}); ok {
						filter = &BookFilter{}
						if genre, ok := f["genre"].(Genre); ok {
							filter.Genre = &genre
						}
						if minRating, ok := f["minRating"].(float64); ok {
							filter.MinRating = &minRating
						}
						if title, ok := f["title"].(string); ok {
							filter.Title = &title
						}
						if authorID, ok := f["authorId"].(string); ok {
							filter.AuthorID = &authorID
						}
					}

					var pagination *PaginationInput
					if pg, ok := p.Args["pagination"].(map[string]interface{}); ok {
						pagination = &PaginationInput{}
						if first, ok := pg["first"].(int); ok {
							pagination.First = &first
						}
						if after, ok := pg["after"].(string); ok {
							pagination.After = &after
						}
					}

					return store.GetBooks(filter, pagination)
				},
			},
			"book": &graphql.Field{
				Type: bookType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, _ := p.Args["id"].(string)
					return store.GetBook(id)
				},
			},
			"authors": &graphql.Field{
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(authorType))),
				Args: graphql.FieldConfigArgument{
					"limit": &graphql.ArgumentConfig{Type: graphql.Int},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					var limit *int
					if l, ok := p.Args["limit"].(int); ok {
						limit = &l
					}
					return store.GetAuthors(limit)
				},
			},
			"author": &graphql.Field{
				Type: authorType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, _ := p.Args["id"].(string)
					return store.GetAuthor(id)
				},
			},
		},
	})

	// Mutation type
	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createBook": &graphql.Field{
				Type: graphql.NewNonNull(bookType),
				Args: graphql.FieldConfigArgument{
					"input": &graphql.ArgumentConfig{Type: graphql.NewNonNull(createBookInput)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					input := p.Args["input"].(map[string]interface{})
					title := input["title"].(string)
					isbn := input["isbn"].(string)
					publishedYear := input["publishedYear"].(int)
					authorID := input["authorId"].(string)
					genre := input["genre"].(Genre)

					var rating *float64
					if r, ok := input["rating"].(float64); ok {
						rating = &r
					}

					return store.CreateBook(title, isbn, publishedYear, authorID, genre, rating)
				},
			},
			"updateBook": &graphql.Field{
				Type: graphql.NewNonNull(bookType),
				Args: graphql.FieldConfigArgument{
					"id":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
					"input": &graphql.ArgumentConfig{Type: graphql.NewNonNull(updateBookInput)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id := p.Args["id"].(string)
					input := p.Args["input"].(map[string]interface{})

					title := ""
					if t, ok := input["title"].(string); ok {
						title = t
					}

					isbn := ""
					if i, ok := input["isbn"].(string); ok {
						isbn = i
					}

					var publishedYear *int
					if py, ok := input["publishedYear"].(int); ok {
						publishedYear = &py
					}

					var rating *float64
					if r, ok := input["rating"].(float64); ok {
						rating = &r
					}

					return store.UpdateBook(id, title, isbn, publishedYear, rating)
				},
			},
			"deleteBook": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Boolean),
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id := p.Args["id"].(string)
					return store.DeleteBook(id)
				},
			},
			"createAuthor": &graphql.Field{
				Type: graphql.NewNonNull(authorType),
				Args: graphql.FieldConfigArgument{
					"input": &graphql.ArgumentConfig{Type: graphql.NewNonNull(createAuthorInput)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					input := p.Args["input"].(map[string]interface{})
					name := input["name"].(string)

					var bio *string
					if b, ok := input["bio"].(string); ok {
						bio = &b
					}

					return store.CreateAuthor(name, bio)
				},
			},
		},
	})

	// Create schema
	return graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	})
}

// HTTP Handler
func graphqlHandler(schema graphql.Schema) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var params struct {
			Query         string                 `json:"query"`
			OperationName string                 `json:"operationName"`
			Variables     map[string]interface{} `json:"variables"`
		}

		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result := graphql.Do(graphql.Params{
			Schema:         schema,
			RequestString:  params.Query,
			VariableValues: params.Variables,
			OperationName:  params.OperationName,
			Context:        r.Context(),
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

func main() {
	store := NewStore()

	// Seed some data
	author1, _ := store.CreateAuthor("Alan Donovan", strPtr("Co-author of The Go Programming Language"))
	author2, _ := store.CreateAuthor("Brian Kernighan", strPtr("Co-author of The Go Programming Language and many other books"))

	rating1 := 4.8
	store.CreateBook("The Go Programming Language", "978-0134190440", 2015, author1.ID, GenreNonfiction, &rating1)

	rating2 := 4.5
	store.CreateBook("The C Programming Language", "978-0131103627", 1988, author2.ID, GenreNonfiction, &rating2)

	schema, err := buildSchema(store)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/graphql", graphqlHandler(schema))

	fmt.Println("GraphQL server running on :8080")
	fmt.Println("Send POST requests to http://localhost:8080/graphql")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func strPtr(s string) *string {
	return &s
}

// ExecuteQuery is a helper for testing
func ExecuteQuery(schema graphql.Schema, query string, variables map[string]interface{}) *graphql.Result {
	return graphql.Do(graphql.Params{
		Schema:         schema,
		RequestString:  query,
		VariableValues: variables,
		Context:        context.Background(),
	})
}
