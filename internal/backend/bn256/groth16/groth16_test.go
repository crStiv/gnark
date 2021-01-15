// Copyright 2020 ConsenSys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by gnark DO NOT EDIT

package groth16_test

import (
	"github.com/consensys/gurvy/bn256/fr"

	curve "github.com/consensys/gurvy/bn256"

	bn256backend "github.com/consensys/gnark/internal/backend/bn256"

	"bytes"
	"github.com/fxamacker/cbor/v2"
	"testing"

	bn256groth16 "github.com/consensys/gnark/internal/backend/bn256/groth16"
	bn256witness "github.com/consensys/gnark/internal/backend/bn256/witness"

	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/r1cs"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/internal/backend/circuits"
	"github.com/consensys/gurvy"
)

func TestCircuits(t *testing.T) {
	for name, circuit := range circuits.Circuits {
		t.Run(name, func(t *testing.T) {
			assert := groth16.NewAssert(t)
			r1cs, err := frontend.Compile(curve.ID, circuit.Circuit)
			assert.NoError(err)
			assert.ProverFailed(r1cs, circuit.Bad)
			assert.ProverSucceeded(r1cs, circuit.Good)
		})
	}
}

func TestParsePublicInput(t *testing.T) {

	expectedNames := [2]string{"data", backend.OneWire}

	inputOneWire := make(map[string]interface{})
	inputOneWire[backend.OneWire] = 3
	if _, err := bn256groth16.ParsePublicInput(expectedNames[:], inputOneWire); err == nil {
		t.Fatal("expected ErrMissingAssigment error")
	}

	missingInput := make(map[string]interface{})
	if _, err := bn256groth16.ParsePublicInput(expectedNames[:], missingInput); err == nil {
		t.Fatal("expected ErrMissingAssigment")
	}

	correctInput := make(map[string]interface{})
	correctInput["data"] = 3
	got, err := bn256groth16.ParsePublicInput(expectedNames[:], correctInput)
	if err != nil {
		t.Fatal(err)
	}

	expected := make([]fr.Element, 2)
	expected[0].SetUint64(3).FromMont()
	expected[1].SetUint64(1).FromMont()
	if len(got) != len(expected) {
		t.Fatal("Unexpected length for assignment")
	}
	for i := 0; i < len(got); i++ {
		if !got[i].Equal(&expected[i]) {
			t.Fatal("error public assignment")
		}
	}

}

//--------------------//
//     benches		  //
//--------------------//

type refCircuit struct {
	nbConstraints int
	X             frontend.Variable
	Y             frontend.Variable `gnark:",public"`
}

func (circuit *refCircuit) Define(curveID gurvy.ID, cs *frontend.ConstraintSystem) error {
	for i := 0; i < circuit.nbConstraints; i++ {
		circuit.X = cs.Mul(circuit.X, circuit.X)
	}
	cs.AssertIsEqual(circuit.X, circuit.Y)
	return nil
}

func referenceCircuit() (r1cs.R1CS, frontend.Witness) {
	const nbConstraints = 40000
	circuit := refCircuit{
		nbConstraints: nbConstraints,
	}
	r1cs, err := frontend.Compile(curve.ID, &circuit)
	if err != nil {
		panic(err)
	}

	var good refCircuit
	good.X.Assign(2)

	// compute expected Y
	var expectedY fr.Element
	expectedY.SetUint64(2)

	for i := 0; i < nbConstraints; i++ {
		expectedY.Mul(&expectedY, &expectedY)
	}

	good.Y.Assign(expectedY)

	return r1cs, &good
}

func TestReferenceCircuit(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	assert := groth16.NewAssert(t)
	r1cs, solution := referenceCircuit()
	assert.ProverSucceeded(r1cs, solution)
}

func BenchmarkSetup(b *testing.B) {
	r1cs, _ := referenceCircuit()

	var pk bn256groth16.ProvingKey
	var vk bn256groth16.VerifyingKey
	b.ResetTimer()

	b.Run("setup", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			bn256groth16.Setup(r1cs.(*bn256backend.R1CS), &pk, &vk)
		}
	})
}

func BenchmarkProver(b *testing.B) {
	r1cs, _solution := referenceCircuit()
	solution, err := bn256witness.Full(_solution)
	if err != nil {
		b.Fatal(err)
	}

	var pk bn256groth16.ProvingKey
	bn256groth16.DummySetup(r1cs.(*bn256backend.R1CS), &pk)

	b.ResetTimer()
	b.Run("prover", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = bn256groth16.Prove(r1cs.(*bn256backend.R1CS), &pk, solution, false)
		}
	})
}

func BenchmarkVerifier(b *testing.B) {
	r1cs, _solution := referenceCircuit()
	solution, err := bn256witness.Public(_solution)
	if err != nil {
		b.Fatal(err)
	}

	var pk bn256groth16.ProvingKey
	var vk bn256groth16.VerifyingKey
	bn256groth16.Setup(r1cs.(*bn256backend.R1CS), &pk, &vk)
	proof, err := bn256groth16.Prove(r1cs.(*bn256backend.R1CS), &pk, solution, false)
	if err != nil {
		panic(err)
	}

	b.ResetTimer()
	b.Run("verifier", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bn256groth16.Verify(proof, &vk, solution)
		}
	})
}

func BenchmarkSerialization(b *testing.B) {
	r1cs, _solution := referenceCircuit()
	solution, err := bn256witness.Full(_solution)
	if err != nil {
		b.Fatal(err)
	}

	var pk bn256groth16.ProvingKey
	var vk bn256groth16.VerifyingKey
	bn256groth16.Setup(r1cs.(*bn256backend.R1CS), &pk, &vk)
	proof, err := bn256groth16.Prove(r1cs.(*bn256backend.R1CS), &pk, solution, false)
	if err != nil {
		panic(err)
	}

	b.ReportAllocs()

	// ---------------------------------------------------------------------------------------------
	// bn256groth16.ProvingKey binary serialization
	b.Run("pk: binary serialization (bn256groth16.ProvingKey)", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			_, _ = pk.WriteTo(&buf)
		}
	})
	b.Run("pk: binary deserialization (bn256groth16.ProvingKey)", func(b *testing.B) {
		var buf bytes.Buffer
		_, _ = pk.WriteTo(&buf)
		var pkReconstructed bn256groth16.ProvingKey
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer(buf.Bytes())
			_, _ = pkReconstructed.ReadFrom(buf)
		}
	})
	{
		var buf bytes.Buffer
		_, _ = pk.WriteTo(&buf)
	}

	// ---------------------------------------------------------------------------------------------
	// bn256groth16.ProvingKey binary serialization (uncompressed)
	b.Run("pk: binary raw serialization (bn256groth16.ProvingKey)", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			_, _ = pk.WriteRawTo(&buf)
		}
	})
	b.Run("pk: binary raw deserialization (bn256groth16.ProvingKey)", func(b *testing.B) {
		var buf bytes.Buffer
		_, _ = pk.WriteRawTo(&buf)
		var pkReconstructed bn256groth16.ProvingKey
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer(buf.Bytes())
			_, _ = pkReconstructed.ReadFrom(buf)
		}
	})
	{
		var buf bytes.Buffer
		_, _ = pk.WriteRawTo(&buf)
	}

	// ---------------------------------------------------------------------------------------------
	// bn256groth16.ProvingKey binary serialization (cbor)
	b.Run("pk: binary cbor serialization (bn256groth16.ProvingKey)", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			enc := cbor.NewEncoder(&buf)
			enc.Encode(&pk)
		}
	})
	b.Run("pk: binary cbor deserialization (bn256groth16.ProvingKey)", func(b *testing.B) {
		var buf bytes.Buffer
		enc := cbor.NewEncoder(&buf)
		enc.Encode(&pk)
		var pkReconstructed bn256groth16.ProvingKey
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer(buf.Bytes())
			dec := cbor.NewDecoder(buf)
			dec.Decode(&pkReconstructed)
		}
	})
	{
		var buf bytes.Buffer
		enc := cbor.NewEncoder(&buf)
		enc.Encode(&pk)
	}

	// ---------------------------------------------------------------------------------------------
	// bn256groth16.Proof binary serialization
	b.Run("proof: binary serialization (bn256groth16.Proof)", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			_, _ = proof.WriteTo(&buf)
		}
	})
	b.Run("proof: binary deserialization (bn256groth16.Proof)", func(b *testing.B) {
		var buf bytes.Buffer
		_, _ = proof.WriteTo(&buf)
		var proofReconstructed bn256groth16.Proof
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer(buf.Bytes())
			_, _ = proofReconstructed.ReadFrom(buf)
		}
	})
	{
		var buf bytes.Buffer
		_, _ = proof.WriteTo(&buf)
	}

	// ---------------------------------------------------------------------------------------------
	// bn256groth16.Proof binary serialization (uncompressed)
	b.Run("proof: binary raw serialization (bn256groth16.Proof)", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			_, _ = proof.WriteRawTo(&buf)
		}
	})
	b.Run("proof: binary raw deserialization (bn256groth16.Proof)", func(b *testing.B) {
		var buf bytes.Buffer
		_, _ = proof.WriteRawTo(&buf)
		var proofReconstructed bn256groth16.Proof
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer(buf.Bytes())
			_, _ = proofReconstructed.ReadFrom(buf)
		}
	})
	{
		var buf bytes.Buffer
		_, _ = proof.WriteRawTo(&buf)
	}

	// ---------------------------------------------------------------------------------------------
	// bn256groth16.Proof binary serialization (cbor)
	b.Run("proof: binary cbor serialization (bn256groth16.Proof)", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			enc := cbor.NewEncoder(&buf)
			enc.Encode(&proof)
		}
	})
	b.Run("proof: binary cbor deserialization (bn256groth16.Proof)", func(b *testing.B) {
		var buf bytes.Buffer
		enc := cbor.NewEncoder(&buf)
		enc.Encode(&proof)
		var proofReconstructed bn256groth16.Proof
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer(buf.Bytes())
			dec := cbor.NewDecoder(buf)
			dec.Decode(&proofReconstructed)
		}
	})
	{
		var buf bytes.Buffer
		enc := cbor.NewEncoder(&buf)
		enc.Encode(&proof)
	}

}
