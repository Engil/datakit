package datakit

import (
	"log"

	p9p "github.com/docker/go-p9p"
	"golang.org/x/net/context"
)

type transaction struct {
	client     *Client
	fromBranch string
	newBranch  string
}

// NewTransaction opens a new transaction originating from fromBranch, named
// newBranch.
func NewTransaction(ctx context.Context, client *Client, fromBranch string, newBranch string) (*transaction, error) {

	err := client.Mkdir(ctx, "branch", fromBranch)
	if err != nil {
		log.Println("Failed to Create branch/", fromBranch, err)
		return nil, err
	}
	err = client.Mkdir(ctx, "branch", fromBranch, "transactions", newBranch)
	if err != nil {
		log.Println("Failed to Create branch/", fromBranch, "/transactions/", newBranch, err)
		return nil, err
	}

	return &transaction{client: client, fromBranch: fromBranch, newBranch: newBranch}, nil
}

func (t *transaction) close(ctx context.Context) {
	// TODO: do we need to clear up unmerged branches?
}

// Abort ensures the update will not be committed.
func (t *transaction) Abort(ctx context.Context) {
	t.close(ctx)
}

// Commit merges the newBranch back into the fromBranch, or fails
func (t *transaction) Commit(ctx context.Context) error {
	path := []string{"branch", t.fromBranch, "transactions", t.newBranch, "ctl"}
	file, err := t.client.Open(ctx, p9p.ORDWR, path...)
	if err != nil {
		log.Println("Failed to Open ctl", err)
		return err
	}
	defer file.Close(ctx)
	_, err = file.Write(ctx, []byte("commit"), 0)
	if err != nil {
		log.Println("Failed to Write ctl", err)
		return err
	}
	return nil
}

// Write updates a key=value pair within the transaction.
func (t *transaction) Write(ctx context.Context, path []string, value string) error {
	p := []string{"branch", t.fromBranch, "transactions", t.newBranch, "rw"}
	for _, dir := range path[0 : len(path)-1] {
		p = append(p, dir)
	}
	err := t.client.Mkdir(ctx, p...)
	if err != nil {
		log.Println("Failed to Mkdir", p)
	}
	p = append(p, path[len(path)-1])
	file, err := t.client.Create(ctx, p...)
	if err != nil {
		log.Println("Failed to Create", p)
		return err
	}
	defer file.Close(ctx)
	_, err = file.Write(ctx, []byte(value), 0)
	if err != nil {
		log.Println("Failed to Write", path, "=", value, ":", err)
		return err
	}
	return nil
}
