package main

import (
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	pb "trevor/proto"
)

const (
	rabbitMQURL  = "amqp://admin:admin@localhost:5672/"
	queueName    = "stars_notification"
	turnDuration = 10 * time.Millisecond
)

type server struct {
	pb.UnsafeOperatorServiceServer
}

var phaseState struct {
	mu               sync.Mutex
	status           pb.PhaseStatus_Status
	message          string
	current_stars    int32
	activateHability bool
	extraMoney       int32
	totalLoot        int32
}

//init: inicializa el estado de la fase
func init() {
	phaseState.status = pb.PhaseStatus_AWAITING_ORDERS
}

//ConfirmaCut: confirma la repartición del botín, revisando si el monto recibido es el correcto
func (s *server) ConfirmCut(ctx context.Context, cutDetails *pb.CutDetails) (*pb.Ack, error) {
	cut := cutDetails.ReceivedCut
	total := cutDetails.Loot + cutDetails.ExtraMoeny

	split := total / 4
	var message string
	if cut == split {
		message = "Justo lo que esperaba"
	} else {
		message = "Justo lo que no esperaba"
	}

	log.Println("Heist successful! Confirming cut to Michael.")
	return &pb.Ack{
		Acknowledged: true,
		Message:      message,
	}, nil
}

//consumeStarNotifications: Recibe la actualización de estrellas con rabbitmq
func consumeStarNotifications() {
	var rabbitMQHOST string
	if os.Getenv("RABBITMQ_HOST") == "" {
		rabbitMQHOST = "192.168.1.6"
	} else {
		rabbitMQHOST = os.Getenv("RABBITMQ_HOST")
	}
	conn, err := amqp.Dial("amqp://admin:admin@" + rabbitMQHOST + ":5673/")
	if err != nil {
		log.Printf("Failed to connect to RabbitMQ: %v", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("Failed to open a channel: %v", err)
		return
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		log.Printf("Failed to declare a queue: %v", err)
		return
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Printf("Failed to register a consumer: %v", err)
		return
	}

	log.Println("Listening for star notifications...")
	for d := range msgs {
		stars, _ := strconv.Atoi(string(d.Body))
		phaseState.mu.Lock()
		if phaseState.status == pb.PhaseStatus_IN_PROGESS {
			phaseState.current_stars = int32(stars)
			log.Printf("<- Received star update: Now at %d stars.", phaseState.current_stars)
			// phaseState.mu.Unlock()
		}
		phaseState.mu.Unlock()
	}
}

//StartDistraction: inicia la distracción, teniendo un 10% de probabilidades de fallar a la mitad de los turnos.
func (s *server) StartDistraction(ctx context.Context, details *pb.DistractionDetails) (*pb.Empty, error) {
	log.Printf("Starting distraction, %d turns needed", details.TurnsNeeded)
	go func() {
		phaseState.mu.Lock()
		phaseState.status = pb.PhaseStatus_IN_PROGESS
		turns_needed := details.TurnsNeeded
		midway_point := turns_needed / 2
		var turn int32
		phaseState.mu.Unlock()
		for turn = 1; turn <= turns_needed; turn++ {
			time.Sleep(turnDuration)
			phaseState.mu.Lock()
			if turn == midway_point && rand.Intn(100) < 10 {
				log.Printf("Distraction failed at turn %d", turn)
				phaseState.status = pb.PhaseStatus_FAILURE
				phaseState.message = "Trevor was too drunk!"
				phaseState.mu.Unlock()
				break
			}
			phaseState.mu.Unlock()
		}
		phaseState.mu.Lock()
		if phaseState.status != pb.PhaseStatus_FAILURE {
			log.Printf("Distraction succeeded after %d turns", turn-1)
			phaseState.status = pb.PhaseStatus_SUCCESS
		}
		phaseState.mu.Unlock()
	}()
	return &pb.Empty{}, nil
}

//RetrieveLoot: se envian los detalles del botin a Michael.
func (s *server) RetrieveLoot(ctx context.Context, empty *pb.Empty) (*pb.LootDetails, error) {
	return &pb.LootDetails{
		Loot:       phaseState.totalLoot - phaseState.extraMoney,
		ExtraMoney: phaseState.extraMoney,
	}, nil
}

//StartHit: inicia el golpe, si las estrellas suben a 5 se activa la habilidad especial. Si se llega a
//			7 estrellas o más, el golpe falla.
func (s *server) StartHit(ctx context.Context, details *pb.HitDetails) (*pb.Empty, error) {
	log.Printf("Starting hit, %d turns needed", details.TurnsNeeded)
	loot := details.Loot
	go consumeStarNotifications()
	go func() {
		phaseState.status = pb.PhaseStatus_IN_PROGESS
		turns_needed := details.TurnsNeeded
		extraMoney := 0
		for turn := 1; turn <= int(turns_needed); turn++ {
			time.Sleep(turnDuration)
			phaseState.mu.Lock()
			log.Printf("Turn %d: Current stars: %d", turn, phaseState.current_stars)
			if phaseState.current_stars >= 5 {
				log.Printf("Activating Trevor rage ability from turn %d", turn)
				phaseState.activateHability = true
			}
			if phaseState.activateHability && phaseState.current_stars >= 7 {
				log.Printf("Hit failed at turn %d due to 7 or more stars", turn)
				phaseState.status = pb.PhaseStatus_FAILURE
				phaseState.message = "Too many stars! The cops arrived!"
				phaseState.current_stars = 0
				phaseState.mu.Unlock()
				break
			}
			phaseState.mu.Unlock()
		}
		phaseState.mu.Lock()
		if phaseState.status != pb.PhaseStatus_FAILURE {
			log.Printf("Hit succeeded after %d turns, extra money earned: $%d", turns_needed, extraMoney)
			phaseState.current_stars = 0
			phaseState.activateHability = false
			phaseState.extraMoney = int32(extraMoney)
			phaseState.status = pb.PhaseStatus_SUCCESS
			phaseState.totalLoot = loot + phaseState.extraMoney

		}
		phaseState.mu.Unlock()
	}()
	return &pb.Empty{}, nil
}

//CheckDistractionStatus: devuelve el estado actual de la distracción.
func (s *server) CheckDistractionStatus(ctx context.Context, details *pb.Empty) (*pb.PhaseStatus, error) {
	phaseState.mu.Lock()
	defer phaseState.mu.Unlock()
	return &pb.PhaseStatus{
		Status:  phaseState.status,
		Message: phaseState.message,
	}, nil
}

//main: inicializa el servidor gRPC y queda a la espera de llamadas.
func main() {
	rand.Seed(time.Now().UnixNano())
	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	grpc_server := grpc.NewServer()
	pb.RegisterOperatorServiceServer(grpc_server, &server{})
	log.Printf("Trevor gRPC server listening on port 50053")
	log.Printf("RabbitMQ HOST: %s", os.Getenv("RABBITMQ_HOST"))
	if err := grpc_server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
