package extensionrepo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"syscall"

	"github.com/pascalandy/nicely/sdk/contract"
	"github.com/pascalandy/nicely/sdk/extension"
)

// release answers Released <tag> <sha> once the published install passes.
func release(e extension.Extension, args []string) result {
	answer, err := releaseTag(e, args)
	if err != nil {
		return failed("release", err)
	}
	return result{action: "release", answer: []string{answer}}
}

func releaseTag(e extension.Extension, args []string) (string, error) {
	if len(args) < 1 || len(args) > 2 || !regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`).MatchString(args[0]) {
		return "", errors.New("usage: just release vX.Y.Z [fixture]")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	env, err := releaseEnv()
	if err != nil {
		return "", err
	}
	root, err := releaseOutput(ctx, env, "", "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	sha, err := releaseOutput(ctx, env, root, "git", "rev-parse", "HEAD^{commit}")
	if err != nil {
		return "", err
	}
	if err := releasePushed(ctx, env, root, sha); err != nil {
		return "", err
	}
	signed, err := releaseSigned(sha)
	if err != nil {
		return "", err
	}
	manifest, err := contract.ParseManifest(e.Manifest, "")
	if err != nil {
		return "", err
	}
	r, err := newReleaseRun(root, args[0], signed)
	if err != nil {
		return "", err
	}
	fmt.Fprintln(os.Stderr, "release: receipt in", filepath.Join(r.dir, "receipt.json"))
	fixture := ""
	if len(args) == 2 && args[1] != "" {
		fixture, err = filepath.Abs(args[1])
		if err != nil {
			return "", err
		}
	}
	err = r.publish(ctx, env, root, manifest, fixture)
	if err != nil {
		r.Failure = err.Error()
	}
	if writeErr := r.write(); writeErr != nil {
		return "", errors.Join(err, writeErr)
	}
	if err != nil {
		return "", err
	}
	return "Released " + r.Tag + " " + r.SHA, nil
}

func (r *releaseReceipt) publish(ctx context.Context, env []string, root string, m contract.Manifest, fixture string) error {
	acceptance := []string{"just", "acceptance"}
	if fixture != "" {
		acceptance = append(acceptance, fixture)
	}
	// A receipt that names just finish proves nothing about the repository,
	// because a hand can write it. So the gate runs again on the signed
	// commit, before acceptance, which can be slow, and before the tag (#419).
	if err := r.require(ctx, env, root, "finish", "just", "finish"); err != nil {
		return err
	}
	if err := r.require(ctx, env, root, "acceptance", acceptance...); err != nil {
		return err
	}
	// Acceptance and finish cannot change the source that the receipt signed.
	if err := releasePushed(ctx, env, root, r.SHA); err != nil {
		return err
	}
	r.hosted(ctx, env, root)
	message := filepath.Join(r.dir, "signoff.json")
	data, err := json.MarshalIndent(r.Signoff, "", "  ")
	if err != nil {
		return err
	}
	// Keep T9's JSON bytes unchanged in the tag, separate from the release receipt.
	if err := os.WriteFile(message, append(data, '\n'), 0o600); err != nil {
		return err
	}
	if err := r.require(ctx, env, root, "tag", "git", "-c", "tag.gpgSign=false", "tag", "--annotate", r.Tag, r.SHA, "--file", message); err != nil {
		return err
	}
	ref := "refs/tags/" + r.Tag
	if err := r.require(ctx, env, root, "push", "git", "push", "origin", ref+":"+ref); err != nil {
		return err
	}
	r.Pushed = true
	if err := r.write(); err != nil {
		return err
	}
	return r.publishedInstall(ctx, env, root, m, fixture)
}

func releasePushed(ctx context.Context, env []string, root, sha string) error {
	head, err := releaseOutput(ctx, env, root, "git", "rev-parse", "HEAD^{commit}")
	if err != nil {
		return err
	}
	status, err := releaseOutput(ctx, env, root, "git", "status", "--porcelain")
	if err != nil {
		return err
	}
	if status != "" || head != sha {
		return errors.New("release needs an unchanged, clean HEAD; commit the intended files and run just signoff")
	}
	if _, err := releaseOutput(ctx, env, root, "git", "fetch", "--no-tags", "origin", "HEAD"); err != nil {
		return err
	}
	pushed, err := releaseOutput(ctx, env, root, "git", "rev-parse", "FETCH_HEAD^{commit}")
	if err != nil {
		return err
	}
	if pushed != sha {
		return errors.New("release SHA is not the pushed default branch of origin; merge and push the signed SHA first")
	}
	return nil
}
