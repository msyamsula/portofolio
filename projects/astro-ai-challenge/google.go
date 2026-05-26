package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/people/v1"
)

type Contact struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type TimeSlot struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type FreeBusyResult struct {
	Email string     `json:"email"`
	Busy  []TimeSlot `json:"busy"`
}

func resolveContacts(ctx context.Context, token *oauth2.Token, oauthCfg *oauth2.Config, names []string) (map[string][]Contact, error) {
	client := oauthCfg.Client(ctx, token)
	srv, err := people.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("people service: %w", err)
	}

	results := make(map[string][]Contact)

	for _, name := range names {
		// If it's already an email, use it directly
		if strings.Contains(name, "@") {
			results[name] = []Contact{{Name: name, Email: name}}
			continue
		}

		var contacts []Contact

		// 1. Try Workspace directory first (most reliable for org)
		dirContacts := searchDirectory(srv, name)
		contacts = append(contacts, dirContacts...)

		// 2. Try personal contacts
		pContacts := searchPersonalContacts(srv, name)
		contacts = append(contacts, pContacts...)

		// 3. If nothing found, try variations: split name, try parts
		if len(contacts) == 0 {
			parts := strings.Fields(name)
			for _, part := range parts {
				if len(part) < 2 {
					continue
				}
				dirContacts = searchDirectory(srv, part)
				contacts = append(contacts, dirContacts...)
				if len(contacts) > 0 {
					break
				}
			}
		}

		// Deduplicate by email
		contacts = deduplicateContacts(contacts)

		results[name] = contacts
	}

	return results, nil
}

func searchDirectory(srv *people.Service, query string) []Contact {
	resp, err := srv.People.SearchDirectoryPeople().
		Query(query).
		ReadMask("names,emailAddresses,nicknames").
		Sources("DIRECTORY_SOURCE_TYPE_DOMAIN_PROFILE").
		PageSize(10).
		Do()
	if err != nil {
		return nil
	}

	var contacts []Contact
	for _, p := range resp.People {
		if len(p.EmailAddresses) == 0 {
			continue
		}
		displayName := extractDisplayName(p)
		contacts = append(contacts, Contact{
			Name:  displayName,
			Email: p.EmailAddresses[0].Value,
		})
	}
	return contacts
}

func searchPersonalContacts(srv *people.Service, query string) []Contact {
	resp, err := srv.People.SearchContacts().
		Query(query).
		ReadMask("names,emailAddresses,nicknames").
		PageSize(10).
		Do()
	if err != nil {
		return nil
	}

	var contacts []Contact
	for _, result := range resp.Results {
		p := result.Person
		if p == nil || len(p.EmailAddresses) == 0 {
			continue
		}
		displayName := extractDisplayName(p)
		contacts = append(contacts, Contact{
			Name:  displayName,
			Email: p.EmailAddresses[0].Value,
		})
	}
	return contacts
}

func extractDisplayName(p *people.Person) string {
	if len(p.Names) > 0 {
		return p.Names[0].DisplayName
	}
	if len(p.Nicknames) > 0 {
		return p.Nicknames[0].Value
	}
	if len(p.EmailAddresses) > 0 {
		return p.EmailAddresses[0].Value
	}
	return ""
}

func deduplicateContacts(contacts []Contact) []Contact {
	seen := make(map[string]bool)
	var result []Contact
	for _, c := range contacts {
		email := strings.ToLower(c.Email)
		if !seen[email] {
			seen[email] = true
			result = append(result, c)
		}
	}
	return result
}

func checkFreeBusy(ctx context.Context, token *oauth2.Token, oauthCfg *oauth2.Config, emails []string, timeMin, timeMax time.Time) ([]FreeBusyResult, error) {
	client := oauthCfg.Client(ctx, token)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("calendar service: %w", err)
	}

	var items []*calendar.FreeBusyRequestItem
	for _, email := range emails {
		items = append(items, &calendar.FreeBusyRequestItem{Id: email})
	}

	req := &calendar.FreeBusyRequest{
		TimeMin:  timeMin.Format(time.RFC3339),
		TimeMax:  timeMax.Format(time.RFC3339),
		Items:    items,
	}

	resp, err := srv.Freebusy.Query(req).Do()
	if err != nil {
		return nil, fmt.Errorf("freebusy query: %w", err)
	}

	var results []FreeBusyResult
	for email, cal := range resp.Calendars {
		fbr := FreeBusyResult{Email: email}
		for _, busy := range cal.Busy {
			fbr.Busy = append(fbr.Busy, TimeSlot{Start: busy.Start, End: busy.End})
		}
		results = append(results, fbr)
	}

	return results, nil
}

func findAvailableSlots(ctx context.Context, token *oauth2.Token, oauthCfg *oauth2.Config, emails []string, durationMinutes int) ([]TimeSlot, error) {
	now := time.Now()
	// Search the next 5 business days
	timeMax := now.AddDate(0, 0, 7)

	fbResults, err := checkFreeBusy(ctx, token, oauthCfg, emails, now, timeMax)
	if err != nil {
		return nil, err
	}

	// Collect all busy periods
	var allBusy []TimeSlot
	for _, fb := range fbResults {
		allBusy = append(allBusy, fb.Busy...)
	}

	// Generate candidate slots (9am-5pm, every 30 min)
	var slots []TimeSlot
	duration := time.Duration(durationMinutes) * time.Minute

	for d := 0; d < 7; d++ {
		day := now.AddDate(0, 0, d)
		if day.Weekday() == time.Saturday || day.Weekday() == time.Sunday {
			continue
		}

		for hour := 9; hour < 17; hour++ {
			for min := 0; min < 60; min += 30 {
				start := time.Date(day.Year(), day.Month(), day.Day(), hour, min, 0, 0, now.Location())
				end := start.Add(duration)

				if start.Before(now) {
					continue
				}
				if end.Hour() > 17 || (end.Hour() == 17 && end.Minute() > 0) {
					continue
				}

				if !conflictsWithBusy(start, end, allBusy) {
					slots = append(slots, TimeSlot{
						Start: start.Format(time.RFC3339),
						End:   end.Format(time.RFC3339),
					})
				}

				if len(slots) >= 5 {
					return slots, nil
				}
			}
		}
	}

	return slots, nil
}

func conflictsWithBusy(start, end time.Time, busy []TimeSlot) bool {
	for _, b := range busy {
		bStart, _ := time.Parse(time.RFC3339, b.Start)
		bEnd, _ := time.Parse(time.RFC3339, b.End)
		if start.Before(bEnd) && end.After(bStart) {
			return true
		}
	}
	return false
}

func createCalendarEvent(ctx context.Context, token *oauth2.Token, oauthCfg *oauth2.Config, title string, attendeeEmails []string, start, end time.Time) (*calendar.Event, error) {
	client := oauthCfg.Client(ctx, token)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("calendar service: %w", err)
	}

	var attendees []*calendar.EventAttendee
	for _, email := range attendeeEmails {
		attendees = append(attendees, &calendar.EventAttendee{Email: email})
	}

	tz := getIANATimezone(start)

	event := &calendar.Event{
		Summary: title,
		Start: &calendar.EventDateTime{
			DateTime: start.Format(time.RFC3339),
			TimeZone: tz,
		},
		End: &calendar.EventDateTime{
			DateTime: end.Format(time.RFC3339),
			TimeZone: tz,
		},
		Attendees: attendees,
	}

	created, err := srv.Events.Insert("primary", event).SendUpdates("all").Do()
	if err != nil {
		return nil, fmt.Errorf("create event: %w", err)
	}

	return created, nil
}

type CalendarEvent struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	Start           string   `json:"start"`
	End             string   `json:"end"`
	Attendees       []string `json:"attendees"`
	Link            string   `json:"link"`
}

func listUpcomingEvents(ctx context.Context, token *oauth2.Token, oauthCfg *oauth2.Config, query string, maxResults int) ([]CalendarEvent, error) {
	client := oauthCfg.Client(ctx, token)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("calendar service: %w", err)
	}

	now := time.Now().Format(time.RFC3339)
	maxTime := time.Now().AddDate(0, 1, 0).Format(time.RFC3339)

	call := srv.Events.List("primary").
		TimeMin(now).
		TimeMax(maxTime).
		SingleEvents(true).
		OrderBy("startTime").
		MaxResults(int64(maxResults))

	if query != "" {
		call = call.Q(query)
	}

	resp, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}

	var events []CalendarEvent
	for _, item := range resp.Items {
		start := ""
		if item.Start != nil {
			start = item.Start.DateTime
			if start == "" {
				start = item.Start.Date
			}
		}
		end := ""
		if item.End != nil {
			end = item.End.DateTime
			if end == "" {
				end = item.End.Date
			}
		}
		var attendees []string
		for _, a := range item.Attendees {
			attendees = append(attendees, a.Email)
		}
		events = append(events, CalendarEvent{
			ID:        item.Id,
			Title:     item.Summary,
			Start:     start,
			End:       end,
			Attendees: attendees,
			Link:      item.HtmlLink,
		})
	}

	return events, nil
}

func listEventsInRange(ctx context.Context, token *oauth2.Token, oauthCfg *oauth2.Config, timeMin, timeMax string) ([]CalendarEvent, error) {
	client := oauthCfg.Client(ctx, token)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("calendar service: %w", err)
	}

	resp, err := srv.Events.List("primary").
		TimeMin(timeMin).
		TimeMax(timeMax).
		SingleEvents(true).
		OrderBy("startTime").
		MaxResults(50).
		Do()
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}

	var events []CalendarEvent
	for _, item := range resp.Items {
		start := ""
		if item.Start != nil {
			start = item.Start.DateTime
			if start == "" {
				start = item.Start.Date
			}
		}
		end := ""
		if item.End != nil {
			end = item.End.DateTime
			if end == "" {
				end = item.End.Date
			}
		}
		var attendees []string
		for _, a := range item.Attendees {
			attendees = append(attendees, a.Email)
		}
		events = append(events, CalendarEvent{
			ID:        item.Id,
			Title:     item.Summary,
			Start:     start,
			End:       end,
			Attendees: attendees,
			Link:      item.HtmlLink,
		})
	}

	return events, nil
}

func updateCalendarEvent(ctx context.Context, token *oauth2.Token, oauthCfg *oauth2.Config, eventID string, title string, attendeeEmails []string, start, end time.Time) (*calendar.Event, error) {
	client := oauthCfg.Client(ctx, token)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("calendar service: %w", err)
	}

	existing, err := srv.Events.Get("primary", eventID).Do()
	if err != nil {
		return nil, fmt.Errorf("get event: %w", err)
	}

	if title != "" {
		existing.Summary = title
	}
	if !start.IsZero() {
		tz := getIANATimezone(start)
		existing.Start = &calendar.EventDateTime{
			DateTime: start.Format(time.RFC3339),
			TimeZone: tz,
		}
		existing.End = &calendar.EventDateTime{
			DateTime: end.Format(time.RFC3339),
			TimeZone: tz,
		}
	}
	if len(attendeeEmails) > 0 {
		var attendees []*calendar.EventAttendee
		for _, email := range attendeeEmails {
			attendees = append(attendees, &calendar.EventAttendee{Email: email})
		}
		existing.Attendees = attendees
	}

	updated, err := srv.Events.Update("primary", eventID, existing).SendUpdates("all").Do()
	if err != nil {
		return nil, fmt.Errorf("update event: %w", err)
	}

	return updated, nil
}

func getUserEmail(ctx context.Context, token *oauth2.Token, oauthCfg *oauth2.Config) (string, string, error) {
	client := oauthCfg.Client(ctx, token)
	srv, err := people.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return "", "", err
	}

	person, err := srv.People.Get("people/me").PersonFields("names,emailAddresses").Do()
	if err != nil {
		return "", "", err
	}

	var email, name string
	if len(person.EmailAddresses) > 0 {
		email = person.EmailAddresses[0].Value
	}
	if len(person.Names) > 0 {
		name = person.Names[0].DisplayName
	}

	return email, name, nil
}

func getIANATimezone(_ time.Time) string {
	return "Asia/Jakarta"
}

func formatDateTime(iso string) string {
	t, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05", iso)
		if err != nil {
			return iso
		}
	}
	return t.Format("Mon, Jan 2 at 3:04 PM")
}

func formatEmails(emails []string) string {
	return strings.Join(emails, ", ")
}
