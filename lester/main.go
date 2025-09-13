package main

import (
	"context"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"bufio"
	"strings"

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
	waitDuration = 10 * time.Second
	turnDuration = 10 * time.Millisecond
)

var stopChan = make(chan bool)
var rejections int32 = 0
var offers []HeistOfferData

// Estructura donde se almacenará la informacion de cada oferta
type HeistOfferData struct {
	Loot int32
	PoliceRisk int32
	TrevorSuccess int32
	FranklinSuccess int32
}

type server struct {
	pb.UnimplementedLesterServiceServer
}

//ConfirmCut: Verifica si el pago recibido es correcto, teniendo en cuenta que si la division no es exacta
//			  el resto del dinero es para lester. Es un método de LesterService.
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

//ProposeHeistOffer: Le propone una oferta a Michael (con probabilidad del 90%), si Michael rechaza 3 veces
//					 entonces espera 10 segundos antes de volver a dar una oferta. Es un método de LesterService.
func (s *server) ProposeHeistOffer(ctx context.Context, empty *pb.Empty) (*pb.HeistOffer, error) {
	if rejections == 3 {
		log.Printf("Michael rejected 3 offers, making him wait %d seconds", waitDuration/time.Second)
		time.Sleep(waitDuration)
		rejections = 0
	}

	if rand.Int31n(100) < 10 {
		return nil, nil
	}
	
	index := rand.Intn(len(offers))

	offer := &pb.HeistOffer{
		Loot:            offers[index].Loot,
		PoliceRisk:      offers[index].PoliceRisk,
		TrevorSuccess:   offers[index].TrevorSuccess,
		FranklinSuccess: offers[index].FranklinSuccess,
	}
	log.Printf("Proposed offer: &{Loot: %d, PoliceRisk: %d, TrevorSuccess: %d, FranklinSuccess: %d}", offer.Loot, offer.PoliceRisk, offer.TrevorSuccess, offer.FranklinSuccess)
	return offer, nil
}

//DecideOnOffer: Se recibe la decisión de Michael de aceptar o rechazar la oferta, se aumenta o reinicia
// 				 el contador de rechazos según corresponda. Es un método de LesterService.	
func (s *server) DecideOnOffer(ctx context.Context, decision *pb.Decision) (*pb.Empty, error) {
	if !decision.Accepted {
		rejections++
	} else {
		rejections = 0
	}
	return &pb.Empty{}, nil
}

//loadOffers: Carga las ofertas desde un archivo CSV, en caso de que la oferta tenga un campo faltante entonces
//			  se le asigna el valor de -1 a ese campo. Es una función auxiliar.
func loadOffers(filePath string) error{
	file, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
	scanner.Scan()
	for scanner.Scan(){
		line := scanner.Text()
		fields:= strings.Split(line, ",")

		var offerData HeistOfferData

		if fields[0] == ""{
			offerData.Loot = -1
		} else {
			loot, err:= strconv.ParseInt(fields[0], 10, 32)
			if err != nil {
    			loot = -1
			}
			offerData.Loot = int32(loot)
		}

		if fields[1] == ""{
			offerData.FranklinSuccess = -1
		} else {
			franklinSuccess, err:= strconv.ParseInt(fields[1], 10, 32)
			if err != nil {
    			franklinSuccess = -1
			}
			offerData.FranklinSuccess = int32(franklinSuccess)
		}

		if fields[2] == ""{
			offerData.TrevorSuccess = -1
		} else {
			trevorSuccess, err:= strconv.ParseInt(fields[2], 10, 32)
			if err != nil {
    			trevorSuccess = -1
			}
			offerData.TrevorSuccess = int32(trevorSuccess)
		}

		if fields[3] == ""{
			offerData.PoliceRisk = -1
		} else {
			policeRisk, err:= strconv.ParseInt(fields[3], 10, 32)
			if err != nil {
    			policeRisk = -1
			}
			offerData.PoliceRisk = int32(policeRisk)
		}

		offers = append(offers, offerData)
	}
	return nil
}

//ManageStarsNotifications: Inicia o detiene el envío de notificaciones de estrellas.
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

//StartStarsNotification: envía actualizaciones de entrellas a rabbit con la frecuencia asignada.
func StartStarsNotification(frequency int) {
	var rabbitMQHOST string
	if os.Getenv("RABBITMQ_HOST") == "" {
		rabbitMQHOST = "192.168.1.6"
	} else {
		rabbitMQHOST = os.Getenv("RABBITMQ_HOST")
	}
	conn, err := amqp.Dial("amqp://admin:admin@" + rabbitMQHOST + ":5673/")

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

//main: Carga las ofertas con loadOffers y luego inicializa el servidor gRPC
func main() {
		err := loadOffers("ofertas_grande.csv")

	if err != nil {
    	log.Fatalf("Failed to load offers from CSV: %v", err)
	}
	
	rand.Seed(time.Now().UnixNano())
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	grpc_server := grpc.NewServer()
	pb.RegisterLesterServiceServer(grpc_server, &server{})
	log.Printf("Lester gRPC server listening on port 50051")
	log.Printf("RabbitMQ HOST: %s", os.Getenv("RABBITMQ_HOST"))
	if err := grpc_server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
