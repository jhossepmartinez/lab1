package main

import (
	"context"
	"log"
	"os"
	// "math/rand"
	// "net"
	"bufio"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "michael/proto"
)

const checkIntervalDuration = 1 * time.Second

//isOfferAcceptable: Checkea si una oferta cumple con los requerimientos de Michael.
func isOfferAcceptable(offer *pb.HeistOffer) bool {
	return (offer.TrevorSuccess > 50 || offer.FranklinSuccess > 50) && offer.PoliceRisk < 80
}

//isOfferValid: Checkea si una oferta tiene todos los campos. La falta de campos es representada con un -1.
func isOfferValid(offer *pb.HeistOffer) bool {
	return (offer.Loot != -1 && offer.TrevorSuccess != -1 && offer.FranklinSuccess != -1 && offer.PoliceRisk != -1)
}

//negotiateOffer: Solicita ofertas a Lester hasta encontrar una oferta que cumpla con sus requerimientos.
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
		
		if !isOfferValid(offer){
			log.Println("Offer has missing fields, rejecting...")
			(*lc).DecideOnOffer(context.Background(), &pb.Decision{Accepted: false})
			continue
		}
		
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

//runDistraction: Asigna a Franklin o Trevor la fase de Distracción en base a sus probabilidades de exito
//				  Luego procede a ejecutar la distracción desde su inicio hasta su finalización.
func runDistraction(trevorClient, franklinClient *pb.OperatorServiceClient, offer *pb.HeistOffer) (*pb.PhaseStatus, string) {
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
			return status, ocName
		}
	}
}

//runHit: Asigna la fase de  Golpe al personaje que no participó en la distracción. Luego se procede a
//		  a ejecutar el golpe desde su inicio hasta su finalización.
func runHit(trevorClient, franklinClient *pb.OperatorServiceClient, offer *pb.HeistOffer) (*pb.PhaseStatus, string) {
	var oc *pb.OperatorServiceClient
	oc = trevorClient
	var ocName string = "Trevor"
	var turns_needed int32 = offer.TrevorSuccess
	if offer.FranklinSuccess < offer.TrevorSuccess {
		ocName = "Franklin"
		oc = franklinClient
		turns_needed = offer.FranklinSuccess
	}
	log.Printf("Running the HIT with %s", ocName)
	_, err := (*oc).StartHit(context.Background(), &pb.HitDetails{TurnsNeeded: 200 - turns_needed, Loot: offer.Loot})
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
			return status, ocName
		}
	}
}

//createReport: Se genera el archivo Reporte.txt con el resumen final de la misión.
func createReport(loot, extraMoney, totalLoot, franklinCut, trevorCut, lesterCut, remainder int32,
	franklinResp, trevorResp, lesterResp string) {
	file, err := os.Create("Reporte.txt")
	if err != nil {
		log.Fatal("Could not create report file: ", err)
	}
	defer file.Close()

	// Format numbers with commas for thousands
	formatNumber := func(num int32) string {
		return fmt.Sprintf("$%d,%03d", num/1000, num%1000)
	}

	// Write the report
	writer := bufio.NewWriter(file)

	writer.WriteString("= = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = =\n")
	writer.WriteString("== REPORTE FINAL DE LA MISION ==\n")
	writer.WriteString("= = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = =\n")
	writer.WriteString("Mision : Asalto al Banco # 7128\n")
	writer.WriteString("Resultado Global : MISION COMPLETADA CON EXITO !\n")
	writer.WriteString("--- REPARTO DEL BOTIN ---\n")
	writer.WriteString(fmt.Sprintf("Botin Base : %s\n", formatNumber(loot)))
	writer.WriteString(fmt.Sprintf("Botin Extra ( Habilidad de Chop ): %s\n", formatNumber(extraMoney)))
	writer.WriteString(fmt.Sprintf("Botin Total : %s\n", formatNumber(totalLoot)))
	writer.WriteString("- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -\n")
	writer.WriteString(fmt.Sprintf("Pago a Franklin : %s\n", formatNumber(franklinCut)))
	writer.WriteString(fmt.Sprintf("Respuesta de Franklin : \"%s\"\n", franklinResp))
	writer.WriteString(fmt.Sprintf("Pago a Trevor : %s\n", formatNumber(trevorCut)))
	writer.WriteString(fmt.Sprintf("Respuesta de Trevor : \"%s\"\n", trevorResp))
	writer.WriteString(fmt.Sprintf("Pago a Lester : %s ( reparto ) + %s ( resto )\n", formatNumber(lesterCut-remainder), formatNumber(remainder)))
	writer.WriteString(fmt.Sprintf("Respuesta de Lester : \"%s\"\n", lesterResp))
	writer.WriteString("- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -\n")
	writer.WriteString(fmt.Sprintf("Saldo Final de la Operacion : %s\n", formatNumber(totalLoot)))
	writer.WriteString("= = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = =\n")

	writer.Flush()

	log.Println("Reporte.txt creado exitosamente")
}

//createReportFailure: Se genera el archivo Reporte.txt al fallar la misión.
func createReportFailure(loot, extraMoney int32, responsable, causa, fase string) {
	file, err := os.Create("Reporte.txt")
	if err != nil {
		log.Fatal("Could not create report file: ", err)
	}
	defer file.Close()
	var totalLoot int32 = loot + extraMoney
	// Format numbers with commas for thousands
	formatNumber := func(num int32) string {
		return fmt.Sprintf("$%d,%03d", num/1000, num%1000)
	}

	// Write the report
	writer := bufio.NewWriter(file)

	writer.WriteString("= = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = =\n")
	writer.WriteString("== REPORTE FINAL DE LA MISION ==\n")
	writer.WriteString("= = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = =\n")
	writer.WriteString("Mision : Asalto al Banco # 7128\n")
	writer.WriteString("Resultado Global : FRACASO ROTUNDO !\n")
	writer.WriteString("--- PERDIDAS ---\n")
	writer.WriteString(fmt.Sprintf("Fase : \"%s\"\n", fase))
	writer.WriteString(fmt.Sprintf("Responsable : %s\n", responsable))
	writer.WriteString(fmt.Sprintf("Causa : %s\n", causa))
	writer.WriteString(fmt.Sprintf("Botin Perdido : %s\n", formatNumber(loot)))
	writer.WriteString(fmt.Sprintf("Botin Extra Perdido ( Habilidad de Chop ): %s\n", formatNumber(extraMoney)))
	writer.WriteString(fmt.Sprintf("Botin Total Perdido: %s\n", formatNumber(totalLoot)))
	writer.WriteString("= = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = = =\n")

	writer.Flush()

	log.Println("Reporte.txt creado exitosamente")
}

//manageLootSplit: Gestiona el reparto de botin, recibe confirmación de los otros 3 personajes y genera el 
//				   reporte con la función createReport().
func manageLootSplit(trevorClient, franklinClient *pb.OperatorServiceClient, lesterClient *pb.LesterServiceClient, ocName string) (int32, int32) {
	var oc *pb.OperatorServiceClient
	oc = trevorClient
	if ocName == "Franklin" {
		oc = franklinClient
	} else {
		oc = trevorClient
	}
	lootDetails, err := (*oc).RetrieveLoot(context.Background(), &pb.Empty{})
	if err != nil {
		log.Fatal("Could not retrieve loot: &v", err)
	}
	loot := lootDetails.Loot
	extraMoney := lootDetails.ExtraMoney
	totalLoot := loot + extraMoney

	split := totalLoot / 4
	remainder := totalLoot % 4

	lesterCut := split + remainder
	franklinCut := split
	trevorCut := split

	ackTrevor, err := (*trevorClient).ConfirmCut(context.Background(), &pb.CutDetails{
		Loot:        loot,
		ExtraMoeny:  extraMoney,
		ReceivedCut: trevorCut,
	})
	if err != nil {
		log.Fatal("Could not retrieve loot: &v", err)
	}
	log.Printf("Trevor's response: %s", ackTrevor.Message)

	ackFranklin, err := (*franklinClient).ConfirmCut(context.Background(), &pb.CutDetails{
		Loot:        loot,
		ExtraMoeny:  extraMoney,
		ReceivedCut: franklinCut,
	})
	if err != nil {
		log.Fatal("Could not retrieve loot: &v", err)
	}
	log.Printf("Franklin's response: %s", ackFranklin.Message)

	ackLester, err := (*lesterClient).ConfirmCut(context.Background(), &pb.CutDetails{
		Loot:        loot,
		ExtraMoeny:  extraMoney,
		ReceivedCut: lesterCut,
	})
	if err != nil {
		log.Fatal("Could not retrieve loot: &v", err)
	}
	log.Printf("Lester's response: %s", ackLester.Message)
	createReport(loot, extraMoney, totalLoot, franklinCut, trevorCut, lesterCut, remainder, ackFranklin.Message, ackTrevor.Message, ackLester.Message)

	return lootDetails.Loot, lootDetails.ExtraMoney
}

//main: Se realiza la conexión con los servicios y se ejecutan las fases del atraco.
func main() {
	localHost := "192.168.1.6"
	lesterHost := os.Getenv("LESTER_HOST")
	trevorHost := os.Getenv("TREVOR_HOST")
	franklinHost := os.Getenv("FRANKLIN_HOST")
	if lesterHost == "" {
		lesterHost = localHost
	}
	if trevorHost == "" {
		trevorHost = localHost
	}
	if franklinHost == "" {
		franklinHost = localHost
	}
	log.Printf("Using hosts - lester: %s, trevor: %s, franklin: %s", lesterHost, trevorHost, franklinHost)
	lesterConn, err := grpc.Dial("10.35.168.23:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to lester: %v", err)
	}
	defer lesterConn.Close()
	lesterClient := pb.NewLesterServiceClient(lesterConn)

	franklinConn, err := grpc.Dial("10.35.168.26:50054", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to franklin: %v", err)
	}
	defer franklinConn.Close()
	franklinClient := pb.NewOperatorServiceClient(franklinConn)

	trevorConn, err := grpc.Dial("10.35.168.25:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
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
	distractionStatus, ocName := runDistraction(&trevorClient, &franklinClient, offer)
	if distractionStatus.Status != pb.PhaseStatus_SUCCESS {
		log.Printf("Coordinating: Phase 2, distraction failed, %s %s", distractionStatus.Message, ocName)
		var causa string
		if ocName == "Trevor" {
			causa = "Trevor estaba borracho"
		} else {
			causa = "Chop ladró y distrajo a Franklin"
		}
		createReportFailure(offer.Loot, 0, ocName, causa, "Distracción")
		return
	}
	log.Println("Coordinating: Phase 2, success")
	log.Println("Coordinating: Phase 3, the hit")
	log.Printf("Starting Lester stars notifications")
	lesterClient.ManageStarsNotifications(context.Background(), &pb.NotificationCommand{
		Command:   pb.NotificationCommand_START,
		Frequency: 100 - offer.PoliceRisk,
	})
	hitStatus, ocName := runHit(&trevorClient, &franklinClient, offer)
	if hitStatus.Status != pb.PhaseStatus_SUCCESS {
		log.Printf("Coordinating: Phase 3, hit failed, %s", hitStatus.Message)
		lesterClient.ManageStarsNotifications(context.Background(), &pb.NotificationCommand{
			Command: pb.NotificationCommand_STOP,
		})
		createReportFailure(offer.Loot, hitStatus.ExtraMoney, ocName, "Se llegó al límite de estrellas", "Robo")
		return
	}
	lesterClient.ManageStarsNotifications(context.Background(), &pb.NotificationCommand{
		Command: pb.NotificationCommand_STOP,
	})
	log.Printf("Hit completed, totalLoot: $%d, extraMoney: $%d", hitStatus.TotalLoot, hitStatus.ExtraMoney)
	log.Println("Coordinating: Phase 3, the hit, success")
	log.Println("Coordinating: Phase 4, managing the loot split")
	loot, extraMoney := manageLootSplit(&trevorClient, &franklinClient, &lesterClient, ocName)
	log.Printf("Loot retrieved: $%d, extraMoney: $%d", loot, extraMoney)
	return

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
