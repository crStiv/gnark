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

package plonkfri

import (
	"crypto/sha256"
	"errors"
	"github.com/consensys/gnark-crypto/ecc/bls24-315/fr"
	"github.com/consensys/gnark-crypto/ecc/bls24-315/fr/fri"
	"math/big"

	bls24_315witness "github.com/consensys/gnark/internal/backend/bls24-315/witness"

	fiatshamir "github.com/consensys/gnark-crypto/fiat-shamir"
)

var ErrInvalidAlgebraicRelation = errors.New("algebraic relation does not hold")

func Verify(proof *Proof, vk *VerifyingKey, publicWitness bls24_315witness.Witness) error {

	// 0 - derive the challenges with Fiat Shamir
	hFunc := sha256.New()
	fs := fiatshamir.NewTranscript(hFunc, "gamma", "beta", "alpha", "zeta")

	dataFiatShamir := make([][]byte, len(publicWitness)+3)
	for i := 0; i < len(publicWitness); i++ {
		dataFiatShamir[i] = make([]byte, len(publicWitness[i]))
		copy(dataFiatShamir[i], publicWitness[i].Marshal())
	}
	dataFiatShamir[len(publicWitness)] = make([]byte, fr.Bytes)
	dataFiatShamir[len(publicWitness)+1] = make([]byte, fr.Bytes)
	dataFiatShamir[len(publicWitness)+2] = make([]byte, fr.Bytes)
	copy(dataFiatShamir[len(publicWitness)], proof.LROpp[0].ID)
	copy(dataFiatShamir[len(publicWitness)+1], proof.LROpp[1].ID)
	copy(dataFiatShamir[len(publicWitness)+2], proof.LROpp[2].ID)

	beta, err := deriveRandomness(&fs, "gamma", dataFiatShamir...)
	if err != nil {
		return err
	}

	gamma, err := deriveRandomness(&fs, "beta", nil)
	if err != nil {
		return err
	}

	alpha, err := deriveRandomness(&fs, "alpha", proof.Zpp.ID)
	if err != nil {
		return err
	}

	// compute the size of the domain of evaluation of the committed polynomial,
	// the opening position. The challenge zeta will be g^{i} where i is the opening
	// position, and g is the generator of the fri domain.
	rho := uint64(fri.GetRho())
	friSize := 2 * rho * vk.Size
	var bFriSize big.Int
	bFriSize.SetInt64(int64(friSize))
	frOpeningPosition, err := deriveRandomness(&fs, "zeta", proof.Hpp[0].ID, proof.Hpp[1].ID, proof.Hpp[2].ID)
	if err != nil {
		return err
	}
	var bOpeningPosition big.Int
	bOpeningPosition.SetBytes(frOpeningPosition.Marshal()).Mod(&bOpeningPosition, &bFriSize)
	openingPosition := bOpeningPosition.Uint64()

	shiftedOpeningPosition := (openingPosition + uint64(2*rho)) % friSize
	err = vk.Iopp.VerifyOpening(shiftedOpeningPosition, proof.OpeningsZmp[1], proof.Zpp)
	if err != nil {
		return err
	}

	// 1 - verify that the commitments are low degree polynomials

	// ql, qr, qm, qo, qkIncomplete
	err = vk.Iopp.VerifyProofOfProximity(vk.Qpp[0])
	if err != nil {
		return err
	}
	err = vk.Iopp.VerifyProofOfProximity(vk.Qpp[1])
	if err != nil {
		return err
	}
	err = vk.Iopp.VerifyProofOfProximity(vk.Qpp[2])
	if err != nil {
		return err
	}
	err = vk.Iopp.VerifyProofOfProximity(vk.Qpp[3])
	if err != nil {
		return err
	}
	err = vk.Iopp.VerifyProofOfProximity(vk.Qpp[4])
	if err != nil {
		return err
	}

	// l, r, o
	err = vk.Iopp.VerifyProofOfProximity(proof.LROpp[0])
	if err != nil {
		return err
	}
	err = vk.Iopp.VerifyProofOfProximity(proof.LROpp[1])
	if err != nil {
		return err
	}
	err = vk.Iopp.VerifyProofOfProximity(proof.LROpp[2])
	if err != nil {
		return err
	}
	err = vk.Iopp.VerifyProofOfProximity(proof.Zpp)
	if err != nil {
		return err
	}

	// h0, h1, h2
	err = vk.Iopp.VerifyProofOfProximity(proof.Hpp[0])
	if err != nil {
		return err
	}
	err = vk.Iopp.VerifyProofOfProximity(proof.Hpp[1])
	if err != nil {
		return err
	}
	err = vk.Iopp.VerifyProofOfProximity(proof.Hpp[2])
	if err != nil {
		return err
	}

	// s1, s2, s3
	err = vk.Iopp.VerifyProofOfProximity(vk.Spp[0])
	if err != nil {
		return err
	}
	err = vk.Iopp.VerifyProofOfProximity(vk.Spp[1])
	if err != nil {
		return err
	}
	err = vk.Iopp.VerifyProofOfProximity(vk.Spp[2])
	if err != nil {
		return err
	}

	// id1, id2, id3
	err = vk.Iopp.VerifyProofOfProximity(vk.Idpp[0])
	if err != nil {
		return err
	}
	err = vk.Iopp.VerifyProofOfProximity(vk.Idpp[1])
	if err != nil {
		return err
	}
	err = vk.Iopp.VerifyProofOfProximity(vk.Idpp[2])
	if err != nil {
		return err
	}

	// Z
	err = vk.Iopp.VerifyProofOfProximity(proof.Zpp)
	if err != nil {
		return err
	}

	// 2 - verify the openings

	// ql, qr, qm, qo, qkIncomplete
	// openingPosition := uint64(2)
	err = vk.Iopp.VerifyOpening(openingPosition, proof.OpeningsQlQrQmQoQkincompletemp[0], vk.Qpp[0])
	if err != nil {
		return err
	}
	err = vk.Iopp.VerifyOpening(openingPosition, proof.OpeningsQlQrQmQoQkincompletemp[1], vk.Qpp[1])
	if err != nil {
		return err
	}
	err = vk.Iopp.VerifyOpening(openingPosition, proof.OpeningsQlQrQmQoQkincompletemp[2], vk.Qpp[2])
	if err != nil {
		return err
	}
	err = vk.Iopp.VerifyOpening(openingPosition, proof.OpeningsQlQrQmQoQkincompletemp[3], vk.Qpp[3])
	if err != nil {
		return err
	}
	err = vk.Iopp.VerifyOpening(openingPosition, proof.OpeningsQlQrQmQoQkincompletemp[4], vk.Qpp[4])
	if err != nil {
		return err
	}

	// l, r, o
	err = vk.Iopp.VerifyOpening(openingPosition, proof.OpeningsLROmp[0], proof.LROpp[0])
	if err != nil {
		return err
	}
	err = vk.Iopp.VerifyOpening(openingPosition, proof.OpeningsLROmp[1], proof.LROpp[1])
	if err != nil {
		return err
	}
	err = vk.Iopp.VerifyOpening(openingPosition, proof.OpeningsLROmp[2], proof.LROpp[2])
	if err != nil {
		return err
	}

	// h0, h1, h2
	err = vk.Iopp.VerifyOpening(openingPosition, proof.OpeningsHmp[0], proof.Hpp[0])
	if err != nil {
		return err
	}
	err = vk.Iopp.VerifyOpening(openingPosition, proof.OpeningsHmp[1], proof.Hpp[1])
	if err != nil {
		return err
	}
	err = vk.Iopp.VerifyOpening(openingPosition, proof.OpeningsHmp[2], proof.Hpp[2])
	if err != nil {
		return err
	}

	// s0, s1, s2
	err = vk.Iopp.VerifyOpening(openingPosition, proof.OpeningsS1S2S3mp[0], vk.Spp[0])
	if err != nil {
		return err
	}
	err = vk.Iopp.VerifyOpening(openingPosition, proof.OpeningsS1S2S3mp[1], vk.Spp[1])
	if err != nil {
		return err
	}
	err = vk.Iopp.VerifyOpening(openingPosition, proof.OpeningsS1S2S3mp[2], vk.Spp[2])
	if err != nil {
		return err
	}

	// id0, id1, id2
	err = vk.Iopp.VerifyOpening(openingPosition, proof.OpeningsId1Id2Id3mp[0], vk.Idpp[0])
	if err != nil {
		return err
	}
	err = vk.Iopp.VerifyOpening(openingPosition, proof.OpeningsId1Id2Id3mp[1], vk.Idpp[1])
	if err != nil {
		return err
	}
	err = vk.Iopp.VerifyOpening(openingPosition, proof.OpeningsId1Id2Id3mp[2], vk.Idpp[2])
	if err != nil {
		return err
	}

	// Z, Zshift
	err = vk.Iopp.VerifyOpening(openingPosition, proof.OpeningsZmp[0], proof.Zpp)
	if err != nil {
		return err
	}

	// verification of the algebraic relation
	var ql, qr, qm, qo, qk fr.Element
	ql.Set(&proof.OpeningsQlQrQmQoQkincompletemp[0].ClaimedValue)
	qr.Set(&proof.OpeningsQlQrQmQoQkincompletemp[1].ClaimedValue)
	qm.Set(&proof.OpeningsQlQrQmQoQkincompletemp[2].ClaimedValue)
	qo.Set(&proof.OpeningsQlQrQmQoQkincompletemp[3].ClaimedValue)
	qk.Set(&proof.OpeningsQlQrQmQoQkincompletemp[4].ClaimedValue) // -> to be completed

	var l, r, o fr.Element
	l.Set(&proof.OpeningsLROmp[0].ClaimedValue)
	r.Set(&proof.OpeningsLROmp[1].ClaimedValue)
	o.Set(&proof.OpeningsLROmp[2].ClaimedValue)

	var h1, h2, h3 fr.Element
	h1.Set(&proof.OpeningsHmp[0].ClaimedValue)
	h2.Set(&proof.OpeningsHmp[1].ClaimedValue)
	h3.Set(&proof.OpeningsHmp[2].ClaimedValue)

	var s1, s2, s3 fr.Element
	s1.Set(&proof.OpeningsS1S2S3mp[0].ClaimedValue)
	s2.Set(&proof.OpeningsS1S2S3mp[1].ClaimedValue)
	s3.Set(&proof.OpeningsS1S2S3mp[2].ClaimedValue)

	var id1, id2, id3 fr.Element
	id1.Set(&proof.OpeningsId1Id2Id3mp[0].ClaimedValue)
	id2.Set(&proof.OpeningsId1Id2Id3mp[1].ClaimedValue)
	id3.Set(&proof.OpeningsId1Id2Id3mp[2].ClaimedValue)

	var z, zshift fr.Element
	z.Set(&proof.OpeningsZmp[0].ClaimedValue)
	zshift.Set(&proof.OpeningsZmp[1].ClaimedValue)

	// 2 - compute the LHS: (ql*l+..+qk)+ α*(z(μx)*(l+β*s₁+γ)*..-z*(l+β*id1+γ))+α²*z*(l1-1)
	// var alpha, beta, gamma fr.Element
	// beta.SetUint64(9)
	// gamma.SetUint64(10)
	// alpha.SetUint64(11)
	var zeta fr.Element
	zeta.Exp(vk.GenOpening, &bOpeningPosition)

	var lhs, t1, t2, t3, tmp, tmp2 fr.Element
	// 2.1 (ql*l+..+qk)
	t1.Mul(&l, &ql)
	tmp.Mul(&r, &qr)
	t1.Add(&t1, &tmp)
	tmp.Mul(&qm, &l).Mul(&tmp, &r)
	t1.Add(&t1, &tmp)
	tmp.Mul(&o, &qo)
	t1.Add(&tmp, &t1)
	tmp = completeQk(publicWitness, vk, zeta)
	tmp.Add(&qk, &tmp)
	t1.Add(&tmp, &t1)

	// 2.2 (z(ux)*(l+β*s1+γ)*..-z*(l+β*id1+γ))
	t2.Mul(&beta, &s1).Add(&t2, &l).Add(&t2, &gamma)
	tmp.Mul(&beta, &s2).Add(&tmp, &r).Add(&tmp, &gamma)
	t2.Mul(&tmp, &t2)
	tmp.Mul(&beta, &s3).Add(&tmp, &o).Add(&tmp, &gamma)
	t2.Mul(&tmp, &t2).Mul(&t2, &zshift)

	tmp.Mul(&beta, &id1).Add(&tmp, &l).Add(&tmp, &gamma)
	tmp2.Mul(&beta, &id2).Add(&tmp2, &r).Add(&tmp2, &gamma)
	tmp.Mul(&tmp, &tmp2)
	tmp2.Mul(&beta, &id3).Add(&tmp2, &o).Add(&tmp2, &gamma)
	tmp.Mul(&tmp2, &tmp).Mul(&tmp, &z)

	t2.Sub(&t2, &tmp)

	// 2.3 (z-1)*l1
	var one fr.Element
	one.SetOne()
	t3.Exp(zeta, big.NewInt(int64(vk.Size))).Sub(&t3, &one)
	tmp.Sub(&zeta, &one).Inverse(&tmp).Mul(&tmp, &vk.SizeInv)
	t3.Mul(&tmp, &t3)
	tmp.Sub(&z, &one)
	t3.Mul(&tmp, &t3)

	// 2.4 (ql*l+s+qk) + α*(z(ux)*(l+β*s1+γ)*...-z*(l+β*id1+γ)..)+ α²*z*(l1-1)
	lhs.Set(&t3).Mul(&lhs, &alpha).Add(&lhs, &t2).Mul(&lhs, &alpha).Add(&lhs, &t1)

	// 3 - compute the RHS
	var rhs fr.Element
	tmp.Exp(zeta, big.NewInt(int64(vk.Size+2)))
	rhs.Mul(&h3, &tmp).
		Add(&rhs, &h2).
		Mul(&rhs, &tmp).
		Add(&rhs, &h1)

	tmp.Exp(zeta, big.NewInt(int64(vk.Size))).Sub(&tmp, &one)
	rhs.Mul(&rhs, &tmp)

	// 4 - verify the relation LHS==RHS
	if !rhs.Equal(&lhs) {
		return ErrInvalidAlgebraicRelation
	}

	return nil

}

// completeQk returns ∑_{i<nb_public_inputs}w_i*L_i
func completeQk(publicWitness bls24_315witness.Witness, vk *VerifyingKey, zeta fr.Element) fr.Element {

	var res fr.Element

	// use L_i+1 = w*Li*(X-z**i)/(X-z**i+1)
	var l, tmp, acc, one fr.Element
	one.SetOne()
	acc.SetOne()
	l.Sub(&zeta, &one).Inverse(&l).Mul(&l, &vk.SizeInv)
	tmp.Exp(zeta, big.NewInt(int64(vk.Size))).Sub(&tmp, &one)
	l.Mul(&l, &tmp)

	for i := 0; i < len(publicWitness); i++ {

		tmp.Mul(&l, &publicWitness[i])
		res.Add(&res, &tmp)

		tmp.Sub(&zeta, &acc)
		l.Mul(&l, &tmp).Mul(&l, &vk.Generator)
		acc.Mul(&acc, &vk.Generator)
		tmp.Sub(&zeta, &acc)
		l.Div(&l, &tmp)
	}

	return res
}
