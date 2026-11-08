// Package orderallocation 釋放一筆訂單佔用的掛單剩餘量與凍結額度，
// 供訂單取消、逾時與後台 refund 共用（獨立成小型 package，避免 order 與 backend_admin 互相依賴）。
package orderallocation

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
	listingrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/listing"
	walletrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/wallet"
)

// ReleaseInTx 依 p2p-order「吃單的凍結規則」釋放 order 佔用的額度：
//
//   - buy 掛單：吃單者（訂單賣方）解凍 cryptoAmount，並還原掛單 remainingAmount。
//   - sell 掛單且掛單為 active/paused：只還原 remainingAmount，額度留在掛單的凍結內。
//   - sell 掛單且掛單已 cancelled/completed：掛單取消時只解凍了當時的 remainingAmount，
//     這筆訂單的額度仍凍結在賣方錢包，因此只解凍賣方，不還原已結束掛單的剩餘量。
//
// 呼叫合約：
//  1. 必須在呼叫端的交易 session 內呼叫，與訂單狀態轉換同一筆交易提交或回滾。
//  2. 只能在訂單的條件式狀態轉換（WHERE status = 預期狀態）影響 1 列之後呼叫；
//     那個條件式 UPDATE 是防止同一筆訂單被重複釋放的唯一守衛。
//  3. 鎖序固定為「訂單列（條件式 UPDATE）→ 掛單列（FOR UPDATE）→ 錢包列」。
func ReleaseInTx(
	ctx context.Context,
	session sqlx.Session,
	listingRepo listingrepo.ListingRepository,
	walletRepo walletrepo.WalletRepository,
	order *entity.Order,
) error {
	listing, err := listingRepo.FindByIDForUpdateInTx(ctx, session, order.ListingID)
	if err != nil {
		return err
	}
	if listing.Type != order.ListingType {
		return fmt.Errorf("order %d listing type %q does not match listing %d type %q", order.ID, order.ListingType, listing.ID, listing.Type)
	}

	switch listing.Type {
	case "buy":
		if err := walletRepo.UnfreezeInTx(ctx, session, order.SellerID, order.CryptoCurrency, order.CryptoAmount); err != nil {
			return err
		}
		return listingRepo.RestoreAmountInTx(ctx, session, listing.ID, order.CryptoAmount)

	case "sell":
		switch listing.Status {
		case "active", "paused":
			return listingRepo.RestoreAmountInTx(ctx, session, listing.ID, order.CryptoAmount)
		case "cancelled", "completed":
			return walletRepo.UnfreezeInTx(ctx, session, order.SellerID, order.CryptoCurrency, order.CryptoAmount)
		default:
			return fmt.Errorf("listing %d has unknown status %q", listing.ID, listing.Status)
		}

	default:
		return fmt.Errorf("listing %d has unknown type %q", listing.ID, listing.Type)
	}
}
