// Copyright 2019 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bookstore

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

//
// The Service type implements a bookstore server.
// All objects are managed in an in-memory non-persistent store.
//
// server is used to implement Bookstoreserver.
type server struct {
	// shelves are stored in a map keyed by shelf id
	// books are stored in a two level map, keyed first by shelf id and then by book id
	Shelves     map[int64]*Shelf
	Books       map[int64]map[int64]*Book
	LastShelfID int64      // the id of the last shelf that was added
	LastBookID  int64      // the id of the last book that was added
	Mutex       sync.Mutex // global mutex to synchronize service access
}

func (s *server) ListShelves(context.Context, *empty.Empty) (*ListShelvesResponse, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	// copy shelf ids from Shelves map keys
	shelves := make([]*Shelf, 0, len(s.Shelves))
	for _, shelf := range s.Shelves {
		shelves = append(shelves, shelf)
	}
	response := &ListShelvesResponse{
		Shelves: shelves,
	}
	return response, nil
}

func (s *server) CreateShelf(ctx context.Context, parameters *CreateShelfParameters) (*Shelf, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	// assign an id and name to a shelf and add it to the Shelves map.
	shelf := parameters.Shelf
	s.LastShelfID++
	sid := s.LastShelfID
	s.Shelves[sid] = shelf

	return shelf, nil

}

func (s *server) DeleteShelves(context.Context, *empty.Empty) (*empty.Empty, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	// delete everything by reinitializing the Shelves and Books maps.
	s.Shelves = make(map[int64]*Shelf)
	s.Books = make(map[int64]map[int64]*Book)
	s.LastShelfID = 0
	s.LastBookID = 0
	return nil, nil
}

func (s *server) GetShelf(ctx context.Context, parameters *GetShelfParameters) (*Shelf, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	// look up a shelf from the Shelves map.
	shelf, err := s.getShelf(parameters.Shelf)
	if err != nil {
		return nil, err
	}

	return shelf, nil
}

func (s *server) DeleteShelf(ctx context.Context, parameters *DeleteShelfParameters) (*empty.Empty, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	// delete a shelf by removing the shelf from the Shelves map and the associated books from the Books map.
	delete(s.Shelves, parameters.Shelf)
	delete(s.Books, parameters.Shelf)
	return nil, nil
}

func (s *server) ListBooks(ctx context.Context, parameters *ListBooksParameters) (responses *ListBooksResponse, err error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	// list the books in a shelf
	_, err = s.getShelf(parameters.Shelf)
	if err != nil {
		return nil, err
	}
	shelfBooks := s.Books[parameters.Shelf]
	books := make([]*Book, 0, len(shelfBooks))
	for _, book := range shelfBooks {
		books = append(books, book)
	}

	response := &ListBooksResponse{
		Books: books,
	}
	return response, nil
}

func (s *server) CreateBook(ctx context.Context, parameters *CreateBookParameters) (*Book, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	_, err := s.getShelf(parameters.Shelf)
	if err != nil {
		return nil, err
	}
	// assign an id and name to a book and add it to the Books map.
	s.LastBookID++
	bid := s.LastBookID
	book := parameters.Book
	if s.Books[parameters.Shelf] == nil {
		s.Books[parameters.Shelf] = make(map[int64]*Book)
	}
	s.Books[parameters.Shelf][bid] = book

	return book, nil
}

func (s *server) GetBook(ctx context.Context, parameters *GetBookParameters) (*Book, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	// get a book from the Books map
	book, err := s.getBook(parameters.Shelf, parameters.Book)
	if err != nil {
		return nil, err
	}

	return book, nil
}

func (s *server) DeleteBook(ctx context.Context, parameters *DeleteBookParameters) (*empty.Empty, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	// delete a book by removing the book from the Books map.
	delete(s.Books[parameters.Shelf], parameters.Book)
	return nil, nil
}

// internal helpers
func (s *server) getShelf(sid int64) (shelf *Shelf, err error) {
	shelf, ok := s.Shelves[sid]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Couldn't find shelf %d", sid))
	} else {
		return shelf, nil
	}
}

func (s *server) getBook(sid int64, bid int64) (book *Book, err error) {
	_, err = s.getShelf(sid)
	if err != nil {
		return nil, err
	}
	book, ok := s.Books[sid][bid]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Couldn't find book %d on shelf %d", bid, sid))
	} else {
		return book, nil
	}
}

func RunServer() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	fmt.Printf("\nServer listening on port %v \n", port)
	RegisterBookstoreServer(s, &server{
		Shelves: map[int64]*Shelf{},
		Books:   map[int64]map[int64]*Book{},
	})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
