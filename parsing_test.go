package main

import (
	"os"
	"testing"

	"github.com/gorilla/feeds"
)

func createExpectedFeedForSmallRss() *feeds.Feed {
	feed := &feeds.Feed{}
	feed.Items = []*feeds.Item{}

	feed.Items = append(
		feed.Items,
		&feeds.Item{
			Title:       "SecFloat: Accurate Floating-Point meets Secure 2-Party Computation, by Deevashwer Rathee and Anwesh Bhattacharya and Rahul Sharma and Divya Gupta and Nishanth Chandran and Aseem Rastogi",
			Link:        &feeds.Link{Href: "https://eprint.iacr.org/2022/322"},
			Description: "We build a library SecFloat for secure 2-party computation (2PC) of 32-bit single-precision floating-point operations and math functions. The existing functionalities used in cryptographic works are imprecise and the precise functionalities used in standard libraries are not crypto-friendly, i.e., they use operations that are cheap on CPUs but have exorbitant cost in 2PC. SecFloat bridges this gap with its novel crypto-friendly precise functionalities. Compared to the prior cryptographic libraries, SecFloat is up to six orders of magnitude more precise and up to two orders of magnitude more efficient. Furthermore, against a precise 2PC baseline, SecFloat is three orders of magnitude more efficient. The high precision of SecFloat leads to the first accurate implementation of secure inference. All prior works on secure inference of deep neural networks rely on ad hoc float-to-fixed converters. We evaluate a model where the fixed-point approximations used in privacy-preserving machine learning completely fail and floating-point is necessary. Thus, emphasizing the need for libraries like SecFloat.",
			Id:          "https://eprint.iacr.org/2022/322",
		},
	)
	feed.Items = append(
		feed.Items,
		&feeds.Item{
			Title:       "Provable security of CFB mode of operation with external re-keying, by Vadim Tsypyschev and Iliya Morgasov",
			Link:        &feeds.Link{Href: "https://eprint.iacr.org/2022/291"},
			Description: "In this article it is investigated security of the cipher feedback mode of operation with regularexternal serial re-keying aiming to construct lightweight pseudo-random sequences generator.For this purpose it was introduced new mode of operation called Multi-key CFB, MCFB, andwas obtained the estimations of provable security of this new mode in the LOR-CPA model.Besides that. it was obtained counterexample to well-known result of Abdalla\u0015Bellare aboutsecurity of encryption scheme with external re-keying.",
			Id:          "https://eprint.iacr.org/2022/291",
		},
	)

	return feed
}

func TestParsing(t *testing.T) {

	smallRss, err := os.ReadFile("3rdparty/small-rss.xml")
	if err != nil {
		t.Fatal(err)
	}

	feed, err := parseEprintFeed(smallRss)
	if err != nil {
		t.Fatal(err)
	}

	expectedFeed := createExpectedFeedForSmallRss()

	// Create JSON representations for easy comparision
	feedJson, err := feed.ToJSON()
	if err != nil {
		t.Fatal(err)
	}
	expectedFeedJson, err := expectedFeed.ToJSON()
	if err != nil {
		t.Fatal(err)
	}
	if feedJson != expectedFeedJson {
		t.Fatalf("failed to parse small-rss.xml\nexpected:\n%s\ngot:\n%s", expectedFeedJson, feedJson)
	}
}
