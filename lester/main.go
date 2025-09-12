package main

import (
	"context"
	"log"
	"math/rand"
	"net"
	"strconv"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"

	// "sync"
	"time"

	pb "lester/proto"
)

// amqp://admin:admin@10.35.168.23:5673/

const (
	rabbitMQURL  = "amqp://admin:admin@localhost:5672/"
	queueName    = "stars_notification"
	waitDuration = 5 * time.Second
	turnDuration = 10 * time.Millisecond
)

var stopChan = make(chan bool)
var rejections int32 = 0

type server struct {
	pb.UnimplementedLesterServiceServer
}

func (s *server) ConfirmCut(ctx context.Context, cutDetails *pb.CutDetails) (*pb.Ack, error) {
	cut := cutDetails.ReceivedCut
	total := cutDetails.Loot + cutDetails.ExtraMoeny
	remainder := total % 4

	split := total / 4
	lesterCut := split + remainder
	var message string
	if cut == lesterCut {
		message = "Excelente! el pago es correcto!"
	} else {
		message = "Mal ahi..."
	}

	log.Println("Heist successful! Confirming cut to Michael.")
	return &pb.Ack{
		Acknowledged: true,
		Message:      message,
	}, nil
}
func (s *server) ProposeHeistOffer(ctx context.Context, empty *pb.Empty) (*pb.HeistOffer, error) {
	if rand.Int31n(100) < 10 {
		return nil, nil
	}
	if rejections == 3 {
		log.Printf("Michael rejected 3 offers, making him wait %d seconds", waitDuration/time.Second)
		time.Sleep(waitDuration)
		rejections = 0
	}
	offer := &pb.HeistOffer{
		Loot:            int32(rand.Int31n(1000000) + 500000),
		PoliceRisk:      int32(rand.Int31n(100)),
		TrevorSuccess:   int32(rand.Int31n(100)),
		FranklinSuccess: int32(rand.Int31n(100)),
	}
	log.Printf("Proposed offer: &{Loot: %d, PoliceRisk: %d, TrevorSuccess: %d, FranklinSuccess: %d}", offer.Loot, offer.PoliceRisk, offer.TrevorSuccess, offer.FranklinSuccess)
	return offer, nil
}

func (s *server) DecideOnOffer(ctx context.Context, decision *pb.Decision) (*pb.Empty, error) {
	if !decision.Accepted {
		rejections++
	} else {
		rejections = 0
	}
	return &pb.Empty{}, nil
}

func (s *server) ManageStarsNotifications(ctx context.Context, commandDetails *pb.NotificationCommand) (*pb.Empty, error) {
	log.Printf("Received command to %s stars notifications", commandDetails.Command.String())
	if commandDetails.Command == pb.NotificationCommand_START {
		log.Printf("Starting stars notifications with frequency %d turns", commandDetails.Frequency)
		go StartStarsNotification(int(commandDetails.Frequency))
	} else {
		log.Println("Stopping stars notifications")
		stopChan <- true
	}
	return &pb.Empty{}, nil
}

func StartStarsNotification(frequency int) {
	conn, err := amqp.Dial("amqp://guest:guest@192.168.1.6:5673/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, error := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", error)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	stars := 0
	ticker := time.NewTicker(time.Duration(frequency) * turnDuration)
	// ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	log.Println("Stars notifications started")
	for {
		select {
		case <-ticker.C:
			stars++
			log.Printf("-> Sending star update: Now at %d stars.", stars)
			ch.PublishWithContext(context.Background(), "", q.Name, false, false, amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(strconv.Itoa(stars)),
			})
		case <-stopChan:
			return
		}

	}

}

func main() {
	rand.Seed(time.Now().UnixNano())
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	grpc_server := grpc.NewServer()
	pb.RegisterLesterServiceServer(grpc_server, &server{})
	log.Printf("Lester gRPC server listening on port 50051")
	if err := grpc_server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
