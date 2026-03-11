// FUSEは外部クリスタル8MHz、システムクロックを分周なしに設定(E:FF, H:DF, L:EF)

#include <avr/io.h>
#include <avr/interrupt.h>

// --- キャリブレーション用定数 ---
constexpr uint32_t clk_base = F_CPU / 8 + 32; // タイマーのクロック周波数（8分周後の値）+誤差補正
constexpr uint16_t TARGET_SAMPLES =50;

// --- 共有変数 ---
volatile uint32_t SumV = 0;
volatile uint16_t NumV = 0;
volatile uint8_t  DataReady = 0;

volatile uint32_t snapSum = 0;
volatile uint16_t snapNum = 0;

static uint32_t lastSum = 0;
static uint16_t lastNum = 0;

// --- USART送信処理 (PD1:TxD) ---
void USART_SendChar(char c) {
	// UCSRA: USART制御&ステータスレジスタA
	// UDRE: USARTデータレジスタ空きフラグ（1=送信バッファが空いている）
	while (!(UCSRA & (1 << UDRE)));
	// UDR: USARTデータレジスタ（送受信バッファ）
	UDR = c;
}

void USART_num(uint32_t n) {
	char buf[11];
	uint8_t i = 0;
	do {
		buf[i++] = (n % 10) + '0';
		n /= 10;
	} while (n > 0);
	while (i > 0) USART_SendChar(buf[--i]);
}

// --- インプットキャプチャ割り込み (ICP1ピン:PD6) ---
ISR(TIMER1_CAPT_vect) {
	// ICR1: インプットキャプチャレジスタ（ICP1ピンの信号エッジ検出時のTimer1の値）
	uint16_t cap = ICR1;
	// TIFR: タイマー割り込みフラグレジスタ
	// ICF1: インプットキャプチャフラグ（1を書き込むことでクリア）
	TIFR = (1 << ICF1);

	static uint16_t preCap = 0;
	// タイマーオーバーフロー対応：uint16_tの符号なし減算により差分は常に正しく計算される
	// 例：preCap=65533でタイマーが一周してcap=5の場合、diff = 5 - 65533 = 8（正しい）
	auto diff = cap - preCap;
	
	// ノイズ除去フィルタ：パルス間隔が10ms（10000カウント）未満は無視
	// 1MHz/8=125kHzカウントで10000÷1000000=10ms → 100Hz未満の信号や起動時の不安定な値を除外
	if (diff < 10000) return;

	preCap = cap; // 正当なデータだった場合のみ更新

	SumV += diff;
	NumV++;

	if (NumV >= TARGET_SAMPLES) {
		snapSum = SumV;
		snapNum = NumV;
		SumV = 0;
		NumV = 0;
		DataReady = 1;
	}
}

constexpr uint16_t BAUD = 9600;
constexpr uint16_t UART_Baud(uint16_t baudrate){
	return F_CPU/16/baudrate-1;
}

int main(void) {
	// USART初期化 (9600bps @ 8MHz)
	// UBRRH/UBRRL: USARTボーレートレジスタ（上位/下位8ビット）
	// ボーレート設定値 = F_CPU/(16*baudrate) - 1
	UBRRH = 0;
	UBRRL = UART_Baud(BAUD);
	// UCSRB: USART制御&ステータスレジスタB
	// TXEN: 送信機能有効化ビット
	UCSRB = (1 << TXEN); // enable TxD
	// UCSRC: USART制御&ステータスレジスタC
	// UCSZ1, UCSZ0: データビット長設定（11=8ビット）
	// 8N1 = 8データビット、パリティなし、1ストップビット
	UCSRC = (1 << UCSZ1) | (1 << UCSZ0); // 8N1

	// Timer1初期化
	// TCCR1A: Timer1制御レジスタA（波形生成モード等）
	TCCR1A = 0;
	// TCCR1B: Timer1制御レジスタB（クロック設定、入力キャプチャ設定）
	// ICNC1 (ビット7): ノイズキャンセラ有効（入力信号の4サンプル一致で確定）
	// ICES1 (ビット6): エッジ選択（1=立ち上がりエッジ、0=立ち下がりエッジ）
	// CS11  (ビット1): クロック選択（CS12:CS11:CS10 = 010 で8分周）
	TCCR1B = (1 << ICNC1) | (1 << ICES1) | (1 << CS11);
	// TIMSK: タイマー割り込みマスクレジスタ
	// ICIE1: インプットキャプチャ割り込み有効化ビット
	TIMSK = (1 << ICIE1);

	sei();

	while (1) {
		if (DataReady) {
			cli();
			auto currentSum = snapSum;
			auto currentNum = snapNum;
			DataReady = 0;
			sei();
			
			if (lastNum == 0) {
				lastSum = currentSum;
				lastNum = currentNum;
			}

			// 今回の50個 + 前回の50個 = 計100個分で周波数を算出
			auto totalSum = lastSum + currentSum;
			auto totalNum = lastNum + currentNum;

			// auto freq10000 = (uint32_t)((uint64_t)clk_base * 10000 * currentNum / currentSum);
			auto freq10000 = (uint32_t)((uint64_t)clk_base * 10000 * totalNum / totalSum);

			// 送信: "周波数*10000 サンプル数\r\n"
			USART_num(freq10000);
			USART_SendChar(' ');
			USART_num(currentNum);
			USART_SendChar('\r');
			USART_SendChar('\n');

			lastSum = currentSum;
			lastNum = currentNum;
		}
	}
}
