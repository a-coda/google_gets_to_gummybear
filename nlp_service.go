package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"
	"strings"

	"google.golang.org/grpc"

	"github.com/jdkato/prose/v2"
)

const (
	defaultText = "Mary had a little lamb. It's fleece was white as snow. And everywhere that Mary went the lamb was sure to go."
	port = ":50051"
	address = "localhost:50051"
)

type server struct {
	UnimplementedNLPServer
}

func (s *server) Analyze(ctx context.Context, in *Input) (*Analysis, error) {
	log.Printf("Received: %s", *in.Text)
	doc := analyzeThis(in.Text)
	return convertToAnalysis(doc), nil
}

func analyzeThis(text *string) *prose.Document {
	doc, _ := prose.NewDocument(*text)
	return doc
}

func convertToAnalysis(doc *prose.Document) (*Analysis) {
	entities := []*Entity{}
	docEntities := doc.Entities()
	for i, _ := range docEntities {
		entities = append(entities, &Entity{Text: &docEntities[i].Text, Label: &docEntities[i].Label})
	}
	sentences := []*Sentence{}
	docSentences := doc.Sentences()
	for j, _ := range docSentences {
		sentences = append(sentences, &Sentence{Text: &docSentences[j].Text})
	}
	tokens := []*Token{}
	docTokens := doc.Tokens()
	for k, _ := range docTokens {
		tokens = append(tokens, &Token{Text: &docTokens[k].Text, Tag: &docTokens[k].Tag, Label: &docTokens[k].Label})
	}
	return &Analysis{Entities: entities, Sentences: sentences, Tokens: tokens}
}	

func logAnalysis(analysis *Analysis) {
	for _, entity := range analysis.Entities {
		log.Printf("Entity: %s %s", *entity.Text, *entity.Label)
	}
	for _, sentence := range analysis.Sentences {
		log.Printf("Sentence: %s", *sentence.Text)
	}
	for _, token := range analysis.Tokens {
		log.Printf("Token: %s %s %s", *token.Text, *token.Tag, *token.Label)
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("missing mode")
	} else {
		switch mode := os.Args[1]; mode {
		case "client":
			clientMain()
		case "server":
			serverMain()
		default:
			log.Fatalf("unknown mode")
		}
	}
}

func serverMain() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	RegisterNLPServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func clientMain() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := NewNLPClient(conn)

	// Contact the server and print out its response.
	text := defaultText
	if len(os.Args) > 2 {
		text = strings.Join(os.Args[2:], " ")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	log.Printf("Sending: %s", text)
	analysis, err := c.Analyze(ctx, &Input{Text: &text})
	if err != nil {
		log.Fatalf("could not contact: %v", err)
	}
	logAnalysis(analysis)
}
