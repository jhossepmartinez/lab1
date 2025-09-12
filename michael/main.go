package main

import (
	"context"
	"log"
	// "math/rand"
	// "net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "michael/proto"
)

const checkIntervalDuration = 1 * time.Second

func isOfferAcceptable(offer *pb.HeistOffer) bool {
	return (offer.TrevorSuccess > 50 || offer.FranklinSuccess > 50) && offer.PoliceRisk < 80
}
func negotiateOffer(lc *pb.LesterServiceClient) *pb.HeistOffer {
	for {
		offer, err := (*lc).ProposeHeistOffer(context.Background(), &pb.Empty{})
		if err != nil {
			log.Fatal("Coud not get offer from lester: &v", err)
		}
		if offer == nil {
			log.Println("Lester didn't propose an offer, retrying...")
			continue
		}
		log.Printf("Received offer: &{Loot: %d, PoliceRisk: %d, TrevorSuccess: %d, FranklinSuccess: %d}", offer.Loot, offer.PoliceRisk, offer.TrevorSuccess, offer.FranklinSuccess)
		if isOfferAcceptable(offer) {
			log.Println("Offer is acceptable, accepting...")
			(*lc).DecideOnOffer(context.Background(), &pb.Decision{Accepted: true})
			return offer
		} else {
			log.Println("Offer is not acceptable, rejecting...")
			(*lc).DecideOnOffer(context.Background(), &pb.Decision{Accepted: false})
			continue
		}
	}
}
func runDistraction(trevorClient, franklinClient *pb.OperatorServiceClient, offer *pb.HeistOffer) *pb.PhaseStatus {
	var oc *pb.OperatorServiceClient
	oc = trevorClient
	var ocName string = "Trevor"
	var turns_needed int32 = offer.TrevorSuccess
	if offer.FranklinSuccess > offer.TrevorSuccess {
		ocName = "Franklin"
		oc = franklinClient
		turns_needed = offer.FranklinSuccess
	}
	log.Printf("Running distraction with %s", ocName)
	_, err := (*oc).StartDistraction(context.Background(), &pb.DistractionDetails{TurnsNeeded: 200 - turns_needed})
	if err != nil {
		log.Fatal("Could not start distraction: &v", err)
	}
	for {
		time.Sleep(checkIntervalDuration)
		status, err := (*oc).CheckDistractionStatus(context.Background(), &pb.Empty{})
		if err != nil {
			log.Fatal("Could not check distraction status: &v", err)
		}
		if status.Status != pb.PhaseStatus_IN_PROGESS {
			log.Printf("Distraction finished with status: %v", status.Status)
			return status
		}
	}
}
func runHit(trevorClient, franklinClient *pb.OperatorServiceClient, offer *pb.HeistOffer) *pb.PhaseStatus {
	var oc *pb.OperatorServiceClient
	oc = trevorClient
	var ocName string = "Trevor"
	var turns_needed int32 = offer.TrevorSuccess
	if offer.FranklinSuccess < offer.TrevorSuccess {
		ocName = "Franklin"
		oc = franklinClient
		turns_needed = offer.FranklinSuccess
	}
	log.Printf("Running distraction with %s", ocName)
	_, err := (*oc).StartHit(context.Background(), &pb.HitDetails{TurnsNeeded: 200 - turns_needed})
	if err != nil {
		log.Fatal("Could not start hit: &v", err)
	}
	for {
		time.Sleep(checkIntervalDuration)
		status, err := (*oc).CheckDistractionStatus(context.Background(), &pb.Empty{})
		if err != nil {
			log.Fatal("Could not check hit status: &v", err)
		}
		if status.Status != pb.PhaseStatus_IN_PROGESS {
			log.Printf("Hit finished with status: %v", status.Status)
			return status
		}
	}

}

func main() {
	lesterConn, err := grpc.Dial("192.168.1.6:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to lester: %v", err)
	}
	defer lesterConn.Close()
	lesterClient := pb.NewLesterServiceClient(lesterConn)

	franklinConn, err := grpc.Dial("192.168.1.6:50054", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to franklin: %v", err)
	}
	defer franklinConn.Close()
	franklinClient := pb.NewOperatorServiceClient(franklinConn)

	trevorConn, err := grpc.Dial("192.168.1.6:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to Trevor: %v", err)
	}
	defer trevorConn.Close()
	trevorClient := pb.NewOperatorServiceClient(trevorConn)

	log.Println("Coordinating: Phase 1, getting the offer from lester")
	offer := negotiateOffer(&lesterClient)
	log.Println("Coodinationg: Phase 1, success")
	log.Printf("Accepted offer: &{Loot: %d, PoliceRisk: %d, TrevorSuccess: %d, FranklinSuccess: %d}", offer.Loot, offer.PoliceRisk, offer.TrevorSuccess, offer.FranklinSuccess)

	log.Println("Coordinating: Phase 2, running the distraction with Franklin")
	distractionStatus := runDistraction(&trevorClient, &franklinClient, offer)
	if distractionStatus.Status != pb.PhaseStatus_SUCCESS {
		log.Printf("Coordinating: Phase 2, distraction failed, %s", distractionStatus.Message)
		return
	}
	log.Println("Coordinating: Phase 2, success")
	log.Println("Coordinating: Phase 3, the hit")
	log.Printf("Starting Lester stars notifications")
	lesterClient.ManageStarsNotifications(context.Background(), &pb.NotificationCommand{
		Command:   pb.NotificationCommand_START,
		Frequency: 100 - offer.PoliceRisk,
	})
	hitStatus := runHit(&trevorClient, &franklinClient, offer)
	if hitStatus.Status != pb.PhaseStatus_SUCCESS {
		log.Printf("Coordinating: Phase 3, hit failed, %s", hitStatus.Message)
		lesterClient.ManageStarsNotifications(context.Background(), &pb.NotificationCommand{
			Command: pb.NotificationCommand_STOP,
		})
		return
	}
	lesterClient.ManageStarsNotifications(context.Background(), &pb.NotificationCommand{
		Command: pb.NotificationCommand_STOP,
	})
	log.Println("Coordinating: Phase 3, the hit, success")

	// var distractionStatus *pb.PhaseStatus
	// if offer.FranklinSuccess > offer.TrevorSuccess {
	// 	log.Println("Coordinating: Phase 2, running the distraction with Franklin")
	// 	distractionStatus = runDistraction(&franklinClient, offer)
	// } else {
	// 	log.Println("Coordinating: Phase 2, running the distraction with Trevor")
	// 	distractionStatus = runDistraction(&trevorClient, offer)
	// }
	//
	// if distractionStatus.Status != pb.PhaseStatus_SUCCESS {
	// 	log.Printf("Coordinating: Phase 2, distraction failed, %s", distractionStatus.Message)
	// 	return
	// }

}
