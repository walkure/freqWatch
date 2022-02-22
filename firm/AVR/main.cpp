// FUSEは外部クリスタル8MHz、システムクロックを分周なしに設定(E:FF, H:DF, L:EF)

#include <avr/io.h>   // /usr/lib/avr/include/
#include <avr/sfr_defs.h>
#include <avr/interrupt.h>
#include <util/delay.h>
#include <util/atomic.h>
#include <string.h>


/////////////////////////////////////////////////
// 端子定義
constexpr uint8_t HEARTBEAT=PD5;	 // PORT D5 動作確認パルス出力
constexpr uint8_t ICPPIN=PD6; // PORT D6 = ICP
constexpr uint8_t UARTTX=PD1; // PORT D1 = TX


/////////////////////////////////////////////////
// タイマ１関連
// 周期測定ののための変数

volatile uint16_t NumV,CapV,PreCap;
volatile uint32_t SumV;

constexpr uint8_t freq_base = 50;
constexpr uint32_t clk_base = F_CPU / 8;
constexpr uint16_t cap_ceil = clk_base / (freq_base - 5);  // 1000000/55 = 18181
constexpr uint16_t cap_floor = clk_base / (freq_base + 5); // 1000000/65 = 15384

volatile uint16_t _filtered_count;
volatile uint32_t _filtered_sum;

// 割り込みハンドラ
ISR(TIMER1_CAPT_vect){
	TIFR |= (1<<ICF1);
	CapV = (uint16_t)ICR1 - PreCap;       // 差分を求めて周期を算出
	PreCap = (uint16_t)ICR1;
	
	SumV += CapV;
	NumV++;
	
	if(CapV > cap_floor && CapV < cap_ceil){
		_filtered_count ++;
		_filtered_sum += CapV;
	}
	
}

// タイマ１設定
void TIMER1_Init(void){
	// TCC1でICP信号（PD6端子）の周期を計測する
	TCCR1A = 0; //initialize Timer1
	TCCR1B = 0;
	TCNT1 = 0;

	TCCR1B =  0x02;   // 8MHzを8分周
	TIMSK  |= (1 << ICIE1); // キャプチャ割り込み許可
}

void UpdatePeriod(float *ave,float *filtered_average){

	uint32_t tSumV,filtered_sum;
	uint16_t tNumV,filtered_count;
	// 割り込みで更新している変数をコピー＆初期化する
	cli(); //割り込み停止
	tSumV = SumV; tNumV = NumV;
	SumV = 0; NumV = 0;
	
	filtered_count = _filtered_count; filtered_sum = _filtered_sum;
	_filtered_count = 0; _filtered_sum = 0;
	
	sei();   //割り込み再開
	*ave = (float)tSumV / tNumV;
	if(filtered_count > 0){
		*filtered_average = (float)filtered_sum / filtered_count;
	}else{
		*filtered_average = 0.;
	}
}

/////////////////////////////////////////////////
// ＵＡＲＴ関連
constexpr uint16_t BAUD = 38400;


constexpr uint16_t USART_Baud(uint16_t baudrate){
	return F_CPU/16/baudrate-1;
}

uint8_t txbuf[16],txn=0;
volatile uint8_t txp=0;

ISR(USART_UDRE_vect){
	if(txp<txn){
		while ( !(UCSRA & (1<<UDRE)) );
		UDR = txbuf[txp];
		txp++;
	} else {
		UCSRB &= ~_BV(UDRIE);
	}
}

void USART_Init(uint16_t ubrr){
	UBRRH = (uint8_t)(ubrr>>8);
	UBRRL = (uint8_t)ubrr;
	UCSRC = (1<<USBS)|(3<<UCSZ0);
	txp = 0;
	UCSRB = (1<<RXEN)|(1<<TXEN)|(1<<UDRIE);
}

void USART_txt(uint8_t *txt,uint8_t len){
	while(txp<txn);
	cli();
	memcpy(txbuf,txt,len);
	txp=0;
	txn=len;
	UCSRB |= _BV(UDRIE);
	sei();
}

void USART_znum(uint32_t num,bool crlf){
	uint8_t s=1;    // ゼロサプレス中
	uint32_t d=100000;
	uint32_t n=num;
	uint8_t  c,p=0,b[10];
	while(d>0){
		c=n/d; n=n%d;
		d=d/10;
		if(d==0) s=0;
		if(s==0 || c!=0){
			b[p++]=c+0x30;
			s=0;
		}
	}
	if(crlf){
		b[p++]='\r';b[p++]='\n';
	}else{
		b[p++]=' ';
	}
	
	USART_txt(b,p);
}

// 周波数の10000倍を求める
constexpr float freq_root = 10000.* clk_base;
void USART_freq(float *count,bool crlf){
	if(*count > 1){
		float f = freq_root/(*count);  
		USART_znum(f,crlf);
	}else{
		USART_znum(0,crlf);
	}
}

/////////////////////////////////////////////////
// メイン
int main(void)
{
	PORTB  = 0xFF; // 入力設定時のプルアップ有効
	PORTD  = 0xFF; // 入力設定時のプルアップ有効
	DDRD   = _BV(HEARTBEAT);     // PORT D5を出力に
	DDRD  &= ~_BV(ICPPIN); // ICPピンを入力に
	PORTD &= ~_BV(ICPPIN); // ICPピンのプルアップ無効化
	PORTD  = _BV(UARTTX);  // PORT D1を出力に

	TIMER1_Init();
	USART_Init(USART_Baud(BAUD));

	uint8_t skp = 5;   // 最初の何回かはデータを捨てる
	bool flag = false;
	static float filtered_average,average;
	for(;;){
		if(skp == 0){
			UpdatePeriod(&average,&filtered_average);
			USART_freq(&average,false);
			USART_freq(&filtered_average,true);
		}else{
			skp--;
		}
		if(flag){
			flag=false;
			PORTD |= _BV(HEARTBEAT);
		}else{
			flag=true;
			PORTD &= ~_BV(HEARTBEAT);
		}
		_delay_ms(1000);
	}
}