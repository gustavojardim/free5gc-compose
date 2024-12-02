package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/free5gc/util/ueauth"
)

type AkmaKeyStorage struct {
	Kakma []byte
	Akid  string
	Kafs  map[string][]byte
}

var (
	instance *AkmaKeyStorage
	once     sync.Once
)

func GetAkmaStorageInstance() *AkmaKeyStorage {
	once.Do(func() {
		instance = &AkmaKeyStorage{
			Kafs: make(map[string][]byte),
		}
	})
	return instance
}

func (a *AkmaKeyStorage) SetKakma(kakma []byte) {
	a.Kakma = kakma
}

func (a *AkmaKeyStorage) SetAkid(akid string) {
	a.Akid = akid
}

func (a *AkmaKeyStorage) StoreApplicationFunctionKey(afId string, kaf []byte) {
	a.Kafs[afId] = kaf
}

func (a *AkmaKeyStorage) GetApplicationFunctionKey(afId string) ([]byte, error) {
	kaf, exists := a.Kafs[afId]
	if !exists {
		return nil, errors.New("K_AF for the specified AF_ID not found")
	}
	return kaf, nil
}

func main() {
	http.HandleFunc("/naanf-akma/v1/register-anchorkey", AKMARegister)
	http.HandleFunc("/naanf-akma/v1/retrieve-applicationkey", AFKeyRequest)
	http.HandleFunc("/naanf-akma/v1/remove-context", RemoveContext)

	log.Println("Starting server on :8000")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

type AkmaKeyReg struct {
	AkmaKey string `json:"akmaKey"`
	AKId    string `json:"aKId"`
}

func AKMARegister(w http.ResponseWriter, r *http.Request) {
	fmt.Println("FUNCTION CALLED!!!!")

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req AkmaKeyReg
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	akmaKeyStorageInstance := GetAkmaStorageInstance()
	akmaKeyStorageInstance.SetKakma([]byte(req.AkmaKey))
	akmaKeyStorageInstance.SetAkid(req.AKId)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "AKMA registration successful",
	})
}

func RegisterAKMAKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var akmaKeyInfo AkmaKeyInfo
	if err := json.NewDecoder(r.Body).Decode(&akmaKeyInfo); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Process the AKMA key info here
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(akmaKeyInfo)
}

type AfKeyRequest struct {
	AfId string `json:"afId,omitempty"`
}

type AfKeyResponse struct {
	Kaf string `json:"kaf,omitempty"`
}

/*
DerivateApplicationFunctionKey derives an application-specific key (K_AF) from the AKMA key (K_AKMA) based on the AF_ID.
*/
func DerivateApplicationFunctionKey(afId string) ([]byte, error) {
	akmaInstance := GetAkmaStorageInstance()
	if akmaInstance.Kakma == nil {
		return nil, fmt.Errorf("K_AKMA is not set")
	}

	P0_Kaf := []byte(afId)
	Kaf, err := ueauth.GetKDFValue(akmaInstance.Kakma, "82", P0_Kaf, ueauth.KDFLen(P0_Kaf))
	if err != nil {
		return nil, fmt.Errorf("Application Function Key generation failed: %v", err)
	}

	return Kaf, nil
}

func AFKeyRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Println("AFKeyRequest FUNCTION CALLED!!!!")
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var request AfKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.AfId == "" {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	kaf, err := DerivateApplicationFunctionKey(request.AfId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(AfKeyResponse{Kaf: fmt.Sprintf("%x", kaf)})
}

func GetAKMAAPPKeyMaterial(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var keyRequest AkmaAfKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&keyRequest); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Fetch and return AKMA Application Key Material
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AkmaAfKeyData{})
}

func RemoveContext(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var ctxRemove CtxRemove
	if err := json.NewDecoder(r.Body).Decode(&ctxRemove); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Handle the context removal
	w.WriteHeader(http.StatusNoContent)
}
