# Challenge 141: GraphQL Server

**Difficulty:** ⭐⭐⭐⭐ Hard
**Time Estimate:** 60 minutes

## Description

Build a complete GraphQL API server using gqlgen that manages a book library system. This project demonstrates GraphQL schema design, resolver implementation, queries, mutations, and subscriptions with real-time updates.

## Features

- **Schema-First Design**: Define GraphQL schema with types, queries, and mutations
- **Code Generation**: Use gqlgen to auto-generate resolver interfaces
- **CRUD Operations**: Create, read, update, and delete books and authors
- **Nested Resolvers**: Efficient data fetching for related entities
- **Field Resolvers**: Lazy loading of related data
- **Input Validation**: Validate mutation inputs
- **Error Handling**: Proper GraphQL error responses
- **Pagination**: Cursor-based pagination for large datasets
- **Filtering**: Search and filter books by various criteria
- **Context Usage**: Request-scoped data and authentication

## Schema Overview

```graphql
type Query {
  books(filter: BookFilter, pagination: PaginationInput): BookConnection!
  book(id: ID!): Book
  authors(limit: Int): [Author!]!
  author(id: ID!): Author
}

type Mutation {
  createBook(input: CreateBookInput!): Book!
  updateBook(id: ID!, input: UpdateBookInput!): Book!
  deleteBook(id: ID!): Boolean!
  createAuthor(input: CreateAuthorInput!): Author!
}

type Book {
  id: ID!
  title: String!
  isbn: String!
  publishedYear: Int!
  author: Author!
  genre: Genre!
  rating: Float
  createdAt: String!
}

type Author {
  id: ID!
  name: String!
  bio: String
  books: [Book!]!
}

enum Genre {
  FICTION
  NONFICTION
  SCIFI
  FANTASY
  MYSTERY
  ROMANCE
}
```

## Requirements

1. Implement GraphQL schema with types, queries, and mutations
2. Create resolver implementations for all operations
3. Use in-memory data storage (map-based)
4. Implement field resolvers for efficient data fetching
5. Add input validation for mutations
6. Support pagination with cursor-based approach
7. Implement filtering by title, author, genre
8. Handle concurrent access safely (use sync.RWMutex)
9. Return proper GraphQL errors
10. Write comprehensive tests for resolvers

## Example Queries

```graphql
# Query all books with filtering and pagination
query {
  books(
    filter: { genre: FICTION, minRating: 4.0 }
    pagination: { first: 10 }
  ) {
    edges {
      node {
        id
        title
        author {
          name
        }
        rating
      }
      cursor
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}

# Create a new book
mutation {
  createBook(input: {
    title: "The Go Programming Language"
    isbn: "978-0134190440"
    publishedYear: 2015
    authorId: "1"
    genre: NONFICTION
  }) {
    id
    title
    author {
      name
    }
  }
}

# Get a specific book with all details
query {
  book(id: "1") {
    id
    title
    isbn
    publishedYear
    genre
    rating
    author {
      name
      bio
      books {
        title
      }
    }
  }
}
```

## Implementation Notes

Since we can't use actual code generation in this challenge, we'll implement a GraphQL server manually using the `graphql-go/graphql` library with:

1. Schema definitions in Go code
2. Resolver functions for each field
3. Input types and validation
4. In-memory storage with concurrent access control
5. Query execution engine integration

## Learning Objectives

- GraphQL schema design principles
- Resolver pattern implementation
- N+1 query problem and solutions
- DataLoader pattern concepts
- Field-level resolvers
- Input validation in GraphQL
- Error handling in GraphQL context
- Pagination strategies (cursor-based)
- GraphQL vs REST API differences
- Context propagation in resolvers

## Testing Focus

- Test all query resolvers
- Test all mutation resolvers
- Test field resolvers and data loading
- Test pagination and filtering
- Test input validation
- Test concurrent access
- Test error scenarios
- Benchmark resolver performance
