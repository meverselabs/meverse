package imo

import (
	"io"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
)

/**
생성시

필수 설정값
입금 토큰
프로잭트 토큰 가격: 입금 토큰 x개
프로잭트 토큰 총 갯수

선택가능한 설정값
입금 한계값


umlimit 로 생성
입금한 토큰량에 비례해서 가져갈수 있는 토큰 수량이 결정



출금 토큰 컨트랙트에 입금
화이트 리스팅
Genesis Membership NFT 민팅?

2가지 입금 방법

입금 ->
*/

type ImoContractConstruction struct {
	ProjectOwner     common.Address
	PayToken         common.Address
	ProjectToken     common.Address
	ProjectOffering  *amount.Amount
	ProjectRaising   *amount.Amount
	PayLimit         *amount.Amount
	StartBlock       uint32
	EndBlock         uint32
	HarvestFeeFactor uint16 //max 10000
	WhiteListAddress common.Address
	WhiteListGroupId hash.Hash256
}

func (s *ImoContractConstruction) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Address(w, s.ProjectOwner); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.PayToken); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.ProjectToken); err != nil {
		return sum, err
	}
	if sum, err := sw.Amount(w, s.ProjectOffering); err != nil {
		return sum, err
	}
	if sum, err := sw.Amount(w, s.ProjectRaising); err != nil {
		return sum, err
	}
	if sum, err := sw.Amount(w, s.PayLimit); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint32(w, s.StartBlock); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint32(w, s.EndBlock); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint16(w, s.HarvestFeeFactor); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.WhiteListAddress); err != nil {
		return sum, err
	}
	if sum, err := sw.Hash256(w, s.WhiteListGroupId); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *ImoContractConstruction) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Address(r, &s.ProjectOwner); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.PayToken); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.ProjectToken); err != nil {
		return sum, err
	}
	if sum, err := sr.Amount(r, &s.ProjectOffering); err != nil {
		return sum, err
	}
	if sum, err := sr.Amount(r, &s.ProjectRaising); err != nil {
		return sum, err
	}
	if sum, err := sr.Amount(r, &s.PayLimit); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint32(r, &s.StartBlock); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint32(r, &s.EndBlock); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint16(r, &s.HarvestFeeFactor); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.WhiteListAddress); err != nil {
		return sum, err
	}
	if sum, err := sr.Hash256(r, &s.WhiteListGroupId); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}
