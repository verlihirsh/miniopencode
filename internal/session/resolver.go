package session

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"miniopencode/internal/client"
	"miniopencode/internal/config"
)

type Client interface {
	ListSessions(ctx context.Context) ([]client.Session, error)
	ListMessages(ctx context.Context, sessionID string) ([]client.Message, error)
	CreateSession(ctx context.Context, title string) (string, error)
}

type Resolver struct {
	Client Client
	Config config.Config
	Now    func() time.Time
}

func (r Resolver) Resolve(ctx context.Context, defaultSession string) (string, error) {
	if defaultSession == "" {
		return "", fmt.Errorf("no default session configured")
	}
	if defaultSession != "daily" {
		sessions, err := r.Client.ListSessions(ctx)
		if err != nil {
			return "", err
		}
		for _, s := range sessions {
			if s.ID == defaultSession {
				return s.ID, nil
			}
		}
		return r.Client.CreateSession(ctx, defaultSession)
	}
	return r.resolveDaily(ctx)
}

var dailyRe = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})-daily-(\d+)$`)

func (r Resolver) resolveDaily(ctx context.Context) (string, error) {
	now := r.Now
	if now == nil {
		now = time.Now
	}
	today := now()
	datePrefix := today.Format("2006-01-02")

	sessions, err := r.Client.ListSessions(ctx)
	if err != nil {
		return "", err
	}

	titleFormat := r.Config.Session.DailyTitleFormat
	if titleFormat == "" {
		titleFormat = "2006-01-02-daily-%d"
	}
	buildTitle := func(part int) string {
		layout := titleFormat
		if strings.Contains(layout, "2006-01-02") {
			layout = strings.Replace(layout, "2006-01-02", "%s", 1)
		}
		return fmt.Sprintf(layout, datePrefix, part)
	}

	var todays []client.Session
	for _, s := range sessions {
		matches := dailyRe.FindStringSubmatch(s.Title)
		if len(matches) != 3 {
			continue
		}
		if matches[1] != datePrefix {
			continue
		}
		todays = append(todays, client.Session{ID: s.ID, Title: s.Title, Part: atoi(matches[2])})
	}

	if len(todays) == 0 {
		return r.Client.CreateSession(ctx, buildTitle(1))
	}

	sort.Slice(todays, func(i, j int) bool { return todays[i].Part < todays[j].Part })
	latest := todays[len(todays)-1]

	underLimit, err := r.underLimit(ctx, latest.ID)
	if err != nil {
		return "", err
	}
	if underLimit {
		return latest.ID, nil
	}
	nextPart := latest.Part + 1
	return r.Client.CreateSession(ctx, buildTitle(nextPart))
}

func (r Resolver) underLimit(ctx context.Context, sessionID string) (bool, error) {
	msgs, err := r.Client.ListMessages(ctx, sessionID)
	if err != nil {
		return false, err
	}
	var totalTokens, totalMessages int
	for _, m := range msgs {
		totalMessages++
		if m.Tokens != nil {
			totalTokens += m.Tokens.Input + m.Tokens.Output + m.Tokens.Reasoning
		}
	}

	maxTokens := r.Config.Session.DailyMaxTokens
	if maxTokens == 0 {
		maxTokens = 250000
	}
	maxMessages := r.Config.Session.DailyMaxMessages
	if maxMessages == 0 {
		maxMessages = 4000
	}

	if maxTokens > 0 && totalTokens > maxTokens {
		return false, nil
	}
	if maxMessages > 0 && totalMessages > maxMessages {
		return false, nil
	}
	return true, nil
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
