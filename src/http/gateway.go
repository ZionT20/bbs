package http

import (
	"bytes"
	"encoding/json"
	"github.com/skycoin/bbs/src/misc/boo"
	"github.com/skycoin/bbs/src/misc/inform"
	"github.com/skycoin/bbs/src/misc/keys"
	"github.com/skycoin/bbs/src/store"
	"github.com/skycoin/bbs/src/store/object"
	"github.com/skycoin/skycoin/src/cipher"
	"log"
	"net/http"
	"os"
)

// Gateway represents what is exposed to HTTP interface.
type Gateway struct {
	l      *log.Logger
	Access *store.Access
	Quit   chan int
}

func (g *Gateway) host(mux *http.ServeMux) error {
	g.l = inform.NewLogger(true, os.Stdout, "")

	/*
		<<< NODE >>>
	*/

	// Quits the node.
	mux.HandleFunc("/api/node/quit",
		func(w http.ResponseWriter, r *http.Request) {
			g.Quit <- 0
			send(w)(true, nil)
		})

	// Obtains node status. TODO
	mux.HandleFunc("/api/node/stats",
		func(w http.ResponseWriter, r *http.Request) {
			send(w)(true, nil)
		})

	/*
		<<< TOOLS >>>
	*/

	// Generates a seed.
	mux.HandleFunc("/api/tools/new_seed",
		func(w http.ResponseWriter, r *http.Request) {
			send(w)(keys.GenerateSeed())
		})

	// Generates public/private key pair.
	mux.HandleFunc("/api/tools/new_key_pair",
		func(w http.ResponseWriter, r *http.Request) {
			var pk cipher.PubKey
			var sk cipher.SecKey
			seed := r.FormValue("seed")
			switch seed {
			case "":
				pk, sk = cipher.GenerateKeyPair()
			default:
				pk, sk = cipher.GenerateDeterministicKeyPair([]byte(seed))
			}
			send(w)(
				struct {
					PubKey string `json:"public_key"`
					SecKey string `json:"secret_key"`
				}{
					PubKey: pk.Hex(),
					SecKey: sk.Hex(),
				},
				nil,
			)
		})

	/*
		<<< SESSION >>>
	*/

	// Lists all users.
	mux.HandleFunc("/api/session/users/get_all",
		func(w http.ResponseWriter, r *http.Request) {
			send(w)(g.Access.GetUsers(r.Context()))
		})

	// Creates a new user.
	mux.HandleFunc("/api/session/users/new",
		func(w http.ResponseWriter, r *http.Request) {
			send(w)(g.Access.NewUser(r.Context(), &object.NewUserIO{
				Seed:  r.FormValue("seed"),
				Alias: r.FormValue("alias"),
			}))
		})

	// Deletes a user.
	mux.HandleFunc("/api/session/users/delete",
		func(w http.ResponseWriter, r *http.Request) {
			send(w)(g.Access.DeleteUser(r.Context(), r.FormValue("alias")))
		})

	// User login.
	mux.HandleFunc("/api/session/login",
		func(w http.ResponseWriter, r *http.Request) {
			send(w)(g.Access.Login(r.Context(), &object.LoginIO{
				Alias: r.FormValue("alias"),
			}))
		})

	// User logout.
	mux.HandleFunc("/api/session/logout",
		func(w http.ResponseWriter, r *http.Request) {
			send(w)(g.Access.Logout(r.Context()))
		})

	mux.HandleFunc("/api/session/get_info",
		func(w http.ResponseWriter, r *http.Request) {
			send(w)(g.Access.GetSession(r.Context()))
		})

	/*
		<<< CONNECTIONS >>>
	*/

	// Gets all connections.
	mux.HandleFunc("/api/connections/get_all",
		func(w http.ResponseWriter, r *http.Request) {
			send(w)(g.Access.GetConnections(r.Context()))
		})

	// Creates a new connection.
	mux.HandleFunc("/api/connections/new",
		func(w http.ResponseWriter, r *http.Request) {
			send(w)(g.Access.NewConnection(r.Context(), &object.ConnectionIO{
				Address: r.FormValue("address"),
			}))
		})

	// Deletes a connection.
	mux.HandleFunc("/api/connections/delete",
		func(w http.ResponseWriter, r *http.Request) {
			send(w)(g.Access.DeleteConnection(r.Context(), &object.ConnectionIO{
				Address: r.FormValue("address"),
			}))
		})

	/*
		<<< SUBSCRIPTIONS >>>
	*/

	// Gets all subscriptions (non-master and master).
	mux.HandleFunc("/api/subscriptions/get_all",
		func(w http.ResponseWriter, r *http.Request) {
			send(w)(g.Access.GetSubscriptions(r.Context()))
		})

	// Creates a new subscription.
	mux.HandleFunc("/api/subscriptions/new",
		func(w http.ResponseWriter, r *http.Request) {
			send(w)(g.Access.NewSubscription(r.Context(), &object.BoardIO{
				PubKeyStr: r.FormValue("public_key"),
			}))
		})

	// Deletes a subscription.
	mux.HandleFunc("/api/subscriptions/delete",
		func(w http.ResponseWriter, r *http.Request) {
			send(w)(g.Access.DeleteSubscription(r.Context(), &object.BoardIO{
				PubKeyStr: r.FormValue("public_key"),
			}))
		})

	/*
		<<< ADMIN >>>
	*/

	// Exports an entire board root to file.
	mux.HandleFunc("/api/admin/board/export",
		func(w http.ResponseWriter, r *http.Request) {
			send(w)(g.Access.ExportBoard(r.Context(), &object.ExportBoardIO{
				PubKeyStr: r.FormValue("board_public_key"),
				Name:      r.FormValue("file_name"),
			}))
		})

	// Imports an entire board root from file to CXO.
	mux.HandleFunc("/api/admin/board/import",
		func(w http.ResponseWriter, r *http.Request) {
			send(w)(g.Access.ImportBoard(r.Context(), &object.ExportBoardIO{
				PubKeyStr: r.FormValue("board_public_key"),
				Name:      r.FormValue("file_name"),
			}))
		})

	// Lists boards that have been discovered, but not subscribed to.
	mux.HandleFunc("/api/admin/discover/boards",
		func(w http.ResponseWriter, r *http.Request) {
			send(w)(g.Access.GetDiscoveredBoards(r.Context()))
		})

	/*
		<<< CONTENT >>>
	*/

	// Gets a list of boards; remote and master (boards that this node owns).
	mux.HandleFunc("/api/content/get_boards",
		func(w http.ResponseWriter, r *http.Request) {
			send(w)(g.Access.GetBoards(r.Context()))
		})

	// Gets a single board.
	mux.HandleFunc("/api/content/get_board",
		func(w http.ResponseWriter, r *http.Request) {
			send(w)(g.Access.GetBoard(r.Context(), &object.BoardIO{
				PubKeyStr: r.FormValue("board_public_key"),
			}))
		})

	// Creates and hosts a new board on this node.
	mux.HandleFunc("/api/content/new_board",
		func(w http.ResponseWriter, r *http.Request) {
			send(w)(g.Access.NewBoard(r.Context(), &object.NewBoardIO{
				Seed: r.FormValue("seed"),
				Name: r.FormValue("name"),
				Body: r.FormValue("body"),
			}))
		})

	// Deletes a hosted board from this node.
	mux.HandleFunc("/api/content/delete_board",
		func(w http.ResponseWriter, r *http.Request) {
			send(w)(g.Access.DeleteBoard(r.Context(), &object.BoardIO{
				PubKeyStr: r.FormValue("board_public_key"),
			}))
		})

	// Obtains a view of a board including it's children threads.
	mux.HandleFunc("/api/content/get_board_page",
		func(w http.ResponseWriter, r *http.Request) {
			send(w)(g.Access.GetBoardPage(r.Context(), &object.BoardIO{
				PubKeyStr: r.FormValue("board_public_key"),
			}))
		})

	// Submits a new thread on specified board.
	mux.HandleFunc("/api/content/new_thread",
		func(w http.ResponseWriter, r *http.Request) {
			send(w)(g.Access.NewThread(r.Context(), &object.NewThreadIO{
				BoardPubKeyStr: r.FormValue("board_public_key"),
				Name:           r.FormValue("name"),
				Body:           r.FormValue("body"),
			}))
		})

	// Gets a view of a thread including it's children posts.
	mux.HandleFunc("/api/content/get_thread_page",
		func(w http.ResponseWriter, r *http.Request) {
			send(w)(g.Access.GetThreadPage(r.Context(), &object.ThreadIO{
				BoardPubKeyStr: r.FormValue("board_public_key"),
				ThreadRefStr:   r.FormValue("thread_ref"),
			}))
		})

	// Adds a new text post on specified thread.
	mux.HandleFunc("/api/content/new_post",
		func(w http.ResponseWriter, r *http.Request) {
			send(w)(g.Access.NewPost(r.Context(), &object.NewPostIO{
				BoardPubKeyStr: r.FormValue("board_public_key"),
				ThreadRefStr:   r.FormValue("thread_ref"),
				PostRefStr:     r.FormValue("post_ref"), // Optional.
				Name:           r.FormValue("name"),
				Body:           r.FormValue("body"),
				ImagesStr:      r.FormValue("images"), // Optional.
			}))
		})

	/*
		<<< VOTES >>>
	*/

	// Gets a view of following/avoiding of specified user.
	mux.HandleFunc("/api/votes/get_follow_page",
		func(w http.ResponseWriter, r *http.Request) {
			send(w)(g.Access.GetFollowPage(r.Context(), &object.UserIO{
				BoardPubKeyStr: r.FormValue("board_public_key"),
				UserPubKeyStr:  r.FormValue("user_public_key"),
			}))
		})

	// Votes on a specified user.
	mux.HandleFunc("/api/votes/vote_user",
		func(w http.ResponseWriter, r *http.Request) {
			send(w)(g.Access.VoteUser(r.Context(), &object.UserVoteIO{
				BoardPubKeyStr: r.FormValue("board_public_key"),
				UserPubKeyStr:  r.FormValue("user_public_key"),
				ModeStr:        r.FormValue("mode"),
				TagStr:         r.FormValue("tag"),
			}))
		})

	// Votes on a specified thread.
	mux.HandleFunc("/api/votes/vote_thread",
		func(w http.ResponseWriter, r *http.Request) {
			send(w)(g.Access.VoteThread(r.Context(), &object.ThreadVoteIO{
				BoardPubKeyStr: r.FormValue("board_public_key"),
				ThreadRefStr:   r.FormValue("thread_ref"),
				ModeStr:        r.FormValue("mode"),
				TagStr:         r.FormValue("tag"),
			}))
		})

	// Votes on a specified post.
	mux.HandleFunc("/api/votes/vote_post",
		func(w http.ResponseWriter, r *http.Request) {
			send(w)(g.Access.VotePost(r.Context(), &object.PostVoteIO{
				BoardPubKeyStr: r.FormValue("board_public_key"),
				PostRefStr:     r.FormValue("post_ref"),
				ModeStr:        r.FormValue("mode"),
				TagStr:         r.FormValue("tag"),
			}))
		})

	return nil
}

/*
	<<< HELPER FUNCTIONS >>>
*/

type Error struct {
	Type    int    `json:"type"`
	Title   string `json:"title"`
	Details string `json:"details"`
}

type Response struct {
	Okay  bool        `json:"okay"`
	Data  interface{} `json:"data,omitempty"`
	Error *Error      `json:"error,omitempty"`
}

func send(w http.ResponseWriter) func(v interface{}, e error) error {
	return func(v interface{}, e error) error {
		if e != nil {
			return sendErr(w, e)
		}
		return sendOK(w, v)
	}
}

func sendOK(w http.ResponseWriter, v interface{}) error {
	response := Response{Okay: true, Data: v}
	return sendStatus(w, response, http.StatusOK)
}

func sendErr(w http.ResponseWriter, e error) error {
	if e == nil {
		return sendOK(w, true)
	}

	eType := boo.Type(e)
	eTitle := boo.Message(eType)

	var status int

	switch eType {
	case boo.Unknown, boo.Internal:
		status = http.StatusInternalServerError
	case boo.NotAuthorised, boo.NotMaster:
		status = http.StatusUnauthorized
	case boo.NotFound:
		status = http.StatusNotFound
	case boo.AlreadyExists:
		status = http.StatusConflict
	default:
		status = http.StatusBadRequest
	}

	d := e.Error()
	details := string(bytes.ToUpper([]byte{d[0]})) + d[1:] + "."

	response := Response{
		Okay: false,
		Error: &Error{
			Type:    eType,
			Title:   eTitle,
			Details: details,
		},
	}
	return sendStatus(w, response, status)
}

func sendStatus(w http.ResponseWriter, v interface{}, status int) error {
	data, e := json.Marshal(v)
	if e != nil {
		return e
	}
	sendRaw(w, data, status)
	return nil
}

func sendRaw(w http.ResponseWriter, data []byte, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}
