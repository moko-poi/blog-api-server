package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// シンプルな単一メソッドのインターフェース
// 実装が用で、オブジェクト自身がバリデーション責任を持つ
type Validator interface {
	Valid(ctx context.Context) (problems map[string]string)
}

// encode/decodeを一箇所で処理
// ジェネリクスを使用してタイプセーフにレスポンスをエンコード
// 将来的にXML対応など、別フォーマットが必要になった場合の変更点を最小化
func encode[T any](w http.ResponseWriter, r *http.Request, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

// リクエストボディのデコードを一箇所で処理
// ジェネリクスにより型安全性を確保しつつ、コンパイラが型推論してくれる
func decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}

// デコードとバリデーションを組み合わせた関数
// Validatorインターフェースを実装する型のみ受け付けるよう型制約
// バリデーションエラーは別途map[string]stringで返すことで、フィールド単位のエラーメッセージをクライアントに提供可能
func decodeValid[T Validator](r *http.Request) (T, map[string]string, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, nil, fmt.Errorf("decode json: %w", err)
	}

	// バリデーション実行
	if problems := v.Valid(r.Context()); len(problems) > 0 {
		return v, problems, fmt.Errorf("invalid %T: %d problems", v, len(problems))
	}
	return v, nil, nil
}

// 一貫したエラーレスポンス形式を提供
// Problemsフィールドでフィールドレベルのエラーをクライアントに伝達
type ErrorResponse struct {
	Error    string            `json:"error"`
	Problems map[string]string `json:"problems,omitempty"`
}
