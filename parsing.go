package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gorilla/feeds"
	"github.com/microcosm-cc/bluemonday"
)

var LINES_TO_IGNORE = [...]string{
	"<?xml version=\"1.0\" encoding=\"UTF-8\" ?>",
	"<rss version=\"2.0\">",
	"<channel><title>Cryptology ePrint Archive</title>",
	"<link>https://eprint.iacr.org/</link>",
	"<description>Recently modified papers in the IACR Cryptology ePrint Archive</description>",
	"<language>en-us</language>",
	"<webMaster>webmaster@iacr.org</webMaster>",
	"<managingEditor>eprint-admin@iacr.org</managingEditor>",
	"<generator>None of your business</generator>",
	"<ttl>60</ttl>",
	"</channel></rss>",
}

func lineCanBeIgnored(line string) bool {
	for _, ignorable := range LINES_TO_IGNORE {
		if line == ignorable {
			return true
		}
	}
	return false
}

func stripPrefixAndPostfix(line, prefix, postfix string) (string, error) {
	if !strings.Contains(line, prefix) || !strings.Contains(line, postfix) {
		err := fmt.Errorf("expected line with %s and %s, got: %s", prefix, postfix, line)
		return "", err
	}

	prefixHuh := line[:len(prefix)]
	if prefixHuh != prefix {
		return "", fmt.Errorf("unexpected prefix! expected: %q, got: %q", prefix, prefixHuh)
	}
	postfixHuh := line[len(line)-len(postfix):]
	if postfixHuh != postfix {
		return "", fmt.Errorf("unexpected postfix! expected: %q, got: %q", postfix, postfixHuh)
	}

	line = line[len(prefix):]
	line = line[:len(line)-len(postfix)]
	return line, nil
}

func stripCDATATags(line string) string {
	strippedLine := strings.ReplaceAll(line, "<![CDATA[", "")
	strippedLine = strings.ReplaceAll(strippedLine, "]]>", "")
	return strippedLine
}

func parseLastBuildDate(line string) (*time.Time, error) {
	parsedLine, err := stripPrefixAndPostfix(line, "<lastBuildDate>", "</lastBuildDate>")
	if err != nil {
		return nil, err
	}

	const eprintTimeFormat = "Mon, 2 Jan 2006 15:04:05 -0700"
	t, err := time.Parse(eprintTimeFormat, parsedLine)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

func parseLink(line string) (*feeds.Link, error) {
	parsedLine, err := stripPrefixAndPostfix(line, "<link>", "</link>")
	if err != nil {
		return nil, err
	}
	return &feeds.Link{Href: parsedLine}, nil
}

func parseTitle(line string) (string, error) {
	parsedLine, err := stripPrefixAndPostfix(line, "<title>", "</title>")
	if err != nil {
		return "", err
	}
	parsedLine = stripCDATATags(parsedLine)

	p := bluemonday.StrictPolicy()
	parsedLine = p.Sanitize(parsedLine)

	return parsedLine, nil
}

func parseDescription(line string) (string, error) {
	if !strings.Contains(line, "<description>") || !strings.Contains(line, "</description>") {
		err := fmt.Errorf("expected description line, got: %s", line)
		return "", err
	}
	parsedLine, err := stripPrefixAndPostfix(line, "<description>", "</description>")
	if err != nil {
		return "", err
	}
	parsedLine = stripCDATATags(parsedLine)

	p := bluemonday.StrictPolicy()
	parsedLine = p.Sanitize(parsedLine)

	return parsedLine, nil
}

func parseGuid(line string) (string, error) {
	parsedLine, err := stripPrefixAndPostfix(line, "<guid>", "</guid>")
	if err != nil {
		return "", err
	}
	return parsedLine, nil
}

func parseEprintFeed(feedBytes []byte) (*feeds.Feed, error) {
	// Initialize a feed
	feed := &feeds.Feed{}
	feed.Items = []*feeds.Item{}

	// Wrap feedBytes in a scanner so we can read it line-by-line
	scanner := bufio.NewScanner(bytes.NewReader(feedBytes))
	// Now, read line-by-line and try to parse the line in a SUPER HACKY way.
	for scanner.Scan() {
		line := scanner.Text()
		if lineCanBeIgnored(line) {
			continue
		}
		if strings.Contains(line, "<lastBuildDate>") {
			builddate, err := parseLastBuildDate(line)
			if err != nil {
				return nil, fmt.Errorf("failed to parse last build date: %s", err)
			}
			feed.Updated = *builddate
			continue
		}
		if strings.Contains(line, "<item>") {
			// 1. The next line should be the link, parse it
			scanner.Scan()
			linkLine := scanner.Text()
			link, err := parseLink(linkLine)
			if err != nil {
				return nil, fmt.Errorf("failed to parse link: %s", err)
			}

			// 2. The next line should be the title, parse it
			scanner.Scan()
			titleLine := scanner.Text()
			title, err := parseTitle(titleLine)
			if err != nil {
				return nil, fmt.Errorf("failed to parse title for %q: %s", link.Href, err)
			}

			// 3. The next line should be the start of the description, eat till you see </description> then parse it
			scanner.Scan()
			descriptionLines := scanner.Text()
			for !strings.Contains(descriptionLines, "</description>") {
				scanner.Scan()
				descriptionLines += scanner.Text()
			}
			description, err := parseDescription(descriptionLines)
			if err != nil {
				return nil, fmt.Errorf("failed to parse description for %q: %s", link.Href, err)
			}

			// 4. The next line should be the GUID line, parse it
			scanner.Scan()
			guidLine := scanner.Text()
			guid, err := parseGuid(guidLine)
			if err != nil {
				return nil, fmt.Errorf("failed to parse guid for %q: %s", link.Href, err)
			}

			// 5. The next line should be the item end tag, parse it
			scanner.Scan()
			itemEndLine := scanner.Text()
			if itemEndLine != "</item>" {
				return nil, fmt.Errorf("failed to parse item end line for %q: %s", link.Href, err)
			}

			feed.Items = append(
				feed.Items,
				&feeds.Item{
					Title:       title,
					Link:        link,
					Description: description,
					Id:          guid,
				},
			)
			continue
		}
		return nil, fmt.Errorf("failed to parse line: %s", line)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

	return feed, nil
}
