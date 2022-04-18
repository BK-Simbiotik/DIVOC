package main

import (
	"encoding/json"
	"errors"
	"github.com/divoc/api/config"
	"github.com/divoc/api/pkg/models"
	kafkaService "github.com/divoc/api/pkg/services"
	models2 "github.com/divoc/api/swagger_gen/models"
	"github.com/divoc/kernel_library/services"
	"github.com/go-openapi/strfmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"sort"
	"sync"
	"time"
)

const CERTIFICATE_TYPE_V3 = "certifyV3"
const RECONCILIATION_TYPE = "precautionaryUpdatePrev"

type ReconciliationStatus string

const (
	TEMP_ERROR ReconciliationStatus = "tempError"
	ERROR      ReconciliationStatus = "error"
	SUCCESS    ReconciliationStatus = "success"
)

func getDoseFromCertificate(certificateMap map[string]interface{}) int {
	if doseValue, found := certificateMap["dose"]; found {
		if doseValueFloat, ok := doseValue.(float64); ok {
			return int(doseValueFloat)
		}
	}
	if certificateJson, found := certificateMap["certificate"]; found {
		var certificate models.Certificate
		if certificateString, ok := certificateJson.(string); ok {
			if err := json.Unmarshal([]byte(certificateString), &certificate); err == nil {
				return int(certificate.Evidence[0].Dose)
			} else {
				log.Errorf("Error in reading certificate json %+v", err)
			}
		}
	}

	return 0
}

func getDoseWiseCertificates(certificates []interface{}) map[int][]map[string]interface{} {
	doseWiseCertificates := map[int][]map[string]interface{}{}
	for _, certificateObj := range certificates {
		if certificate, ok := certificateObj.(map[string]interface{}); ok {
			if doseValue := getDoseFromCertificate(certificate); doseValue != 0 {
				doseWiseCertificates[doseValue] = append(doseWiseCertificates[doseValue], certificate)
			}
		}
	}
	return doseWiseCertificates
}

func sortCertificatesByCreateAt(certificateArr []interface{}) []interface{} {
	sort.Slice(certificateArr, func(i, j int) bool {
		certificateA := certificateArr[i].(map[string]interface{})
		certificateB := certificateArr[j].(map[string]interface{})
		certificateACreateAt := certificateA["_osCreatedAt"].(string)
		certificateBCreateAt := certificateB["_osCreatedAt"].(string)
		return certificateACreateAt < certificateBCreateAt
	})
	return certificateArr
}

func compareVaccineDates(metaDate string, dbDate string) (bool, error) {
	metaTime, err := time.Parse("2006-01-02T15:04:05.000Z", metaDate)
	if err != nil {
		return true, err
	}
	dbTime, err := time.Parse("2006-01-02T15:04:05.000Z", dbDate)
	if err != nil {
		return true, err
	}
	my, mm, md := metaTime.Date()
	dy, dm, dd := dbTime.Date()
	if my == dy && mm == dm && md == dd {
		return true, nil
	}
	return false, nil
}

func CheckDataConsistence(requestData *models2.CertificationRequestV2MetaVaccinationsItems0, dbData *models2.CertificationRequestV2Vaccination) (bool, error) {
	var err error
	if !strfmt.IsDate(requestData.Date) && !strfmt.IsDateTime(requestData.Date) {
		log.Error("Invalid vaccination date")
		return true, errors.New("invalid vaccination date")
	}
	metaVaccineDateWithTimestamp := requestData.Date
	if strfmt.IsDate(requestData.Date) {
		date, err := time.Parse("2006-01-02", requestData.Date)
		if err != nil {
			return true, err
		}
		metaVaccineDateWithTimestamp = date.Format("2006-01-02T00:00:00.000Z")
	}
	// assuming that none of these fields should be empty. If empty we will not do the data update
	if requestData.Batch == "" || requestData.Dose < 1 {
		log.Info("Required fields are invalid")
		return true, nil
	}
	vaccineDatesMatched, err := compareVaccineDates(metaVaccineDateWithTimestamp, dbData.Date.String())
	if err != nil {
		log.Error(err)
		return true, err
	}
	if vaccineDatesMatched {
		requestData.Date = dbData.Date.String()
	} else {
		requestData.Date = metaVaccineDateWithTimestamp
	}
	if !vaccineDatesMatched || requestData.Batch != dbData.Batch || requestData.Name != dbData.Name || requestData.Manufacturer != dbData.Manufacturer {
		return false, nil
	}
	return true, nil
}

func CreateUpdateRequestObject(certifyMessage *models2.CertificationRequestV2, dbData *models.Certificate,
	metaData *models2.CertificationRequestV2MetaVaccinationsItems0) (*models2.CertificationRequestV2, error) {
	updateReqV2 := new(models2.CertificationRequestV2)
	updateReqV2.PreEnrollmentCode = certifyMessage.PreEnrollmentCode
	updateReqV2.Recipient = certifyMessage.Recipient
	updateVaccination, err := createVaccinationInfo(certifyMessage, metaData, dbData)
	if err != nil {
		return updateReqV2, err
	}
	updateReqV2.Vaccination = updateVaccination
	updateReqV2.Vaccinator = createVaccinatorInfo(dbData)
	updateFacility, err := createFacilityInfo(dbData)
	if err != nil {
		return updateReqV2, err
	}
	updateReqV2.Facility = updateFacility
	updateReqV2.Meta = createMetaInfo(dbData)
	return updateReqV2, nil
}

func createVaccinatorInfo(dbData *models.Certificate) *models2.CertificationRequestV2Vaccinator {
	vaccinator := new(models2.CertificationRequestV2Vaccinator)
	vaccinator.Name = dbData.Evidence[0].Verifier.Name
	return vaccinator
}

func createVaccinationInfo(certifyMessage *models2.CertificationRequestV2, metaData *models2.CertificationRequestV2MetaVaccinationsItems0,
	dbData *models.Certificate) (*models2.CertificationRequestV2Vaccination, error) {
	vaccinationInfo := new(models2.CertificationRequestV2Vaccination)
	vaccinationInfo.Date, _ = strfmt.ParseDateTime(metaData.Date)
	if metaData.Name != "" {
		vaccinationInfo.Name = metaData.Name
	} else {
		vaccinationInfo.Name = dbData.Evidence[0].Vaccine
	}
	vaccinationInfo.Dose = float64(metaData.Dose)
	if totalDoseValue, ok := dbData.Evidence[0].TotalDoses.(float64); ok {
		vaccinationInfo.TotalDoses = totalDoseValue
	} else {
		vaccinationInfo.TotalDoses = certifyMessage.Vaccination.TotalDoses
	}
	vaccinationInfo.Batch = metaData.Batch
	if metaData.Manufacturer != "" {
		vaccinationInfo.Manufacturer = metaData.Manufacturer
	} else {
		vaccinationInfo.Manufacturer = dbData.Evidence[0].Manufacturer
	}
	effectiveStart, terr := time.Parse("2006-01-02", dbData.Evidence[0].EffectiveStart)
	if terr != nil {
		log.Info("error while parsing effectiveStart Date")
		return vaccinationInfo, terr
	}
	vaccinationInfo.EffectiveStart = strfmt.Date(effectiveStart)
	effectiveUntil, terr := time.Parse("2006-01-02", dbData.Evidence[0].EffectiveUntil)
	if terr != nil {
		log.Info("error while parsing effectiveUntilDate")
		return vaccinationInfo, terr
	}
	vaccinationInfo.EffectiveUntil = strfmt.Date(effectiveUntil)
	return vaccinationInfo, nil
}

func createFacilityInfo(dbData *models.Certificate) (*models2.CertificationRequestV2Facility, error) {
	facility := new(models2.CertificationRequestV2Facility)
	facilityInDB := dbData.Evidence[0].Facility
	facility.Name = facilityInDB.Name
	facilityAddress := new(models2.CertificationRequestV2FacilityAddress)
	facilityAddress.AddressLine1 = facilityInDB.Address.StreetAddress
	facilityAddress.AddressLine2 = facilityInDB.Address.StreetAddress2
	facilityAddress.District = facilityInDB.Address.District
	facilityAddress.State = facilityInDB.Address.AddressRegion
	if facilityInDB.Address.PostalCode != "" {
		if facilityPincode, ok := facilityInDB.Address.PostalCode.(float64); ok {
			facilityAddress.Pincode = int64(facilityPincode)
		} else {
			return facility, errors.New("invalid pin code for facility")
		}
	}
	facility.Address = facilityAddress
	return facility, nil
}

func createMetaInfo(dbData *models.Certificate) *models2.CertificationRequestV2Meta {
	meta := new(models2.CertificationRequestV2Meta)
	meta.PreviousCertificateID = dbData.Evidence[0].CertificateId
	return meta
}

func publishCertifyMessage(request []byte) {
	go kafkaService.PublishCertifyMessage(
		request,
		nil,
		nil,
		kafkaService.MessageHeader{CertificateType: CERTIFICATE_TYPE_V3})
}

func publishReconciliationEvent(preEnrollmentCode string, reconStatus string) {
	go kafkaService.PublishProcStatus(models.ProcStatus{
		Date:              time.Now(),
		PreEnrollmentCode: preEnrollmentCode,
		ProcType:          RECONCILIATION_TYPE,
		Status:            reconStatus,
	})
}

func publishReconErrorRequest(request *models2.CertificationRequestV2) {
	if jsonRequestString, err := json.Marshal(request); err == nil {
		go kafkaService.PublishReconErrorRequest(
			jsonRequestString,
			nil,
			nil,
			kafkaService.MessageHeader{CertificateType: kafkaService.CERTIFICATE_TYPE_V3})
	}
}

func getDBVaccinationData(certificate *models.Certificate) *models2.CertificationRequestV2Vaccination {
	dbData := new(models2.CertificationRequestV2Vaccination)
	dbData.Batch = certificate.Evidence[0].Batch
	dbData.Date = strfmt.DateTime(certificate.Evidence[0].Date)
	dbData.Name = certificate.Evidence[0].Vaccine
	dbData.Manufacturer = certificate.Evidence[0].Manufacturer
	return dbData
}

func reconcileData(certifyMessage *models2.CertificationRequestV2) (ReconciliationStatus, error) {
	start := time.Now()
	filter := map[string]interface{}{
		"preEnrollmentCode": map[string]interface{}{
			"eq": certifyMessage.PreEnrollmentCode,
		},
	}
	certificateFromRegistry, err := services.QueryRegistry("VaccinationCertificate", filter)
	if err != nil {
		return TEMP_ERROR, err
	}
	certificates := certificateFromRegistry["VaccinationCertificate"].([]interface{})
	certificates = sortCertificatesByCreateAt(certificates)
	currentDose := int64(certifyMessage.Vaccination.Dose)
	reconciliationStatus := SUCCESS
	if len(certificates) > 0 {
		certificatesByDose := getDoseWiseCertificates(certificates)
		for _, metaVaccinationData := range certifyMessage.Meta.Vaccinations {
			var certificate models.Certificate
			dose := metaVaccinationData.Dose
			if dose >= currentDose {
				continue
			}
			doseCertificates := certificatesByDose[int(dose)]
			if doseCertificates == nil || len(doseCertificates) == 0 {
				continue
			}
			latestDoseCertificate := doseCertificates[len(doseCertificates)-1]
			if err := json.Unmarshal([]byte(latestDoseCertificate["certificate"].(string)), &certificate); err != nil {
				log.Errorf("Unable to parse certificate string %+v", err)
				return ERROR, err
			}
			dbVaccinationData := getDBVaccinationData(&certificate)
			isDataConsistent, err := CheckDataConsistence(metaVaccinationData, dbVaccinationData)
			if err != nil {
				log.Errorf("Error while checking data consistency %v", err)
				return ERROR, err
			}
			reconciliationStatus, err = publishUpdateRequestForInconsistentData(isDataConsistent, certifyMessage, &certificate, metaVaccinationData)
			if err != nil {
				return reconciliationStatus, err
			}
		}
	}
	log.Infof("Reconciled: %v", time.Since(start))
	return reconciliationStatus, nil
}

func publishUpdateRequestForInconsistentData(isDataConsistent bool, certifyMessage *models2.CertificationRequestV2,
	certificate *models.Certificate, metaVaccinationData *models2.CertificationRequestV2MetaVaccinationsItems0) (ReconciliationStatus, error) {
	if !isDataConsistent {
		updateRequestObject, err := CreateUpdateRequestObject(certifyMessage, certificate, metaVaccinationData)
		if err != nil {
			return ERROR, err
		}
		if jsonRequestString, err := json.Marshal(updateRequestObject); err == nil {
			publishCertifyMessage(jsonRequestString)
			return SUCCESS, err
		} else {
			return ERROR, err
		}
	}
	return SUCCESS, nil
}

func initializeKafka(servers string) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": servers})
	if err != nil {
		panic(err)
	}

	log.Infof("Connected to kafka on %s", servers)

	kafkaService.StartCertifyProducer(producer)
	kafkaService.StartProcStatusEventProducer(producer)
	kafkaService.StartReconErrorRequestProducer(producer)
}

func main() {
	config.Initialize()
	servers := config.Config.Kafka.BootstrapServers
	initializeKafka(servers)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers":  servers,
			"group.id":           "certificate_reconciliation",
			"auto.offset.reset":  "latest",
			"enable.auto.commit": "false",
		})
		if err != nil {
			log.Errorf("error while creating consumer %+v", err)
			panic(err)
		}
		err = consumer.SubscribeTopics([]string{"certify"}, nil)
		if err != nil {
			log.Errorf("error while subscribing to consumer %+v", err)
			panic(err)
		}

		for {
			msg, err := consumer.ReadMessage(-1)
			if err == nil {
				var message models2.CertificationRequestV2
				if err := json.Unmarshal(msg.Value, &message); err == nil {
					if message.Meta != nil && message.Meta.Vaccinations != nil && len(message.Meta.Vaccinations) != 0 {
						status, err := reconcileData(&message)
						publishReconciliationEvent(message.PreEnrollmentCode, string(status))
						if err != nil && status == TEMP_ERROR {
							continue
						} else if status == ERROR {
							publishReconErrorRequest(&message)
						}
					}
				} else {
					log.Errorf("Error unmarshaling certify message %s", err)
				}
				consumer.CommitMessage(msg)
			} else {
				// The client will automatically try to recover from all errors.
				log.Errorf("Consumer error: %v \n", err)
			}
		}
	}()
	wg.Wait()
}