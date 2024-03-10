package huffman

type Symbol struct {
	Code uint32
	Bits uint8
}

var Table = [0xff + 1]Symbol{
	{0b1111111111000, 13},
	{0b11111111111111111011000, 23},
	{0b1111111111111111111111100010, 28},
	{0b1111111111111111111111100011, 28},
	{0b1111111111111111111111100100, 28},
	{0b1111111111111111111111100101, 28},
	{0b1111111111111111111111100110, 28},
	{0b1111111111111111111111100111, 28},
	{0b1111111111111111111111101000, 28},
	{0b111111111111111111101010, 24},
	{0b111111111111111111111111111100, 30},
	{0b1111111111111111111111101001, 28},
	{0b1111111111111111111111101010, 28},
	{0b111111111111111111111111111101, 30},
	{0b1111111111111111111111101011, 28},
	{0b1111111111111111111111101100, 28},
	{0b1111111111111111111111101101, 28},
	{0b1111111111111111111111101110, 28},
	{0b1111111111111111111111101111, 28},
	{0b1111111111111111111111110000, 28},
	{0b1111111111111111111111110001, 28},
	{0b1111111111111111111111110010, 28},
	{0b111111111111111111111111111110, 30},
	{0b1111111111111111111111110011, 28},
	{0b1111111111111111111111110100, 28},
	{0b1111111111111111111111110101, 28},
	{0b1111111111111111111111110110, 28},
	{0b1111111111111111111111110111, 28},
	{0b1111111111111111111111111000, 28},
	{0b1111111111111111111111111001, 28},
	{0b1111111111111111111111111010, 28},
	{0b1111111111111111111111111011, 28},
	{0b010100, 6},
	{0b1111111000, 10},
	{0b1111111001, 10},
	{0b111111111010, 12},
	{0b1111111111001, 13},
	{0b010101, 6},
	{0b11111000, 8},
	{0b11111111010, 11},
	{0b1111111010, 10},
	{0b1111111011, 10},
	{0b11111001, 8},
	{0b11111111011, 11},
	{0b11111010, 8},
	{0b010110, 6},
	{0b010111, 6},
	{0b011000, 6},
	{0b00000, 5},
	{0b00001, 5},
	{0b00010, 5},
	{0b011001, 6},
	{0b011010, 6},
	{0b011011, 6},
	{0b011100, 6},
	{0b011101, 6},
	{0b011110, 6},
	{0b011111, 6},
	{0b1011100, 7},
	{0b11111011, 8},
	{0b111111111111100, 15},
	{0b100000, 6},
	{0b111111111011, 12},
	{0b1111111100, 10},
	{0b1111111111010, 13},
	{0b100001, 6},
	{0b1011101, 7},
	{0b1011110, 7},
	{0b1011111, 7},
	{0b1100000, 7},
	{0b1100001, 7},
	{0b1100010, 7},
	{0b1100011, 7},
	{0b1100100, 7},
	{0b1100101, 7},
	{0b1100110, 7},
	{0b1100111, 7},
	{0b1101000, 7},
	{0b1101001, 7},
	{0b1101010, 7},
	{0b1101011, 7},
	{0b1101100, 7},
	{0b1101101, 7},
	{0b1101110, 7},
	{0b1101111, 7},
	{0b1110000, 7},
	{0b1110001, 7},
	{0b1110010, 7},
	{0b11111100, 8},
	{0b1110011, 7},
	{0b11111101, 8},
	{0b1111111111011, 13},
	{0b1111111111111110000, 19},
	{0b1111111111100, 13},
	{0b11111111111100, 14},
	{0b100010, 6},
	{0b111111111111101, 15},
	{0b00011, 5},
	{0b100011, 6},
	{0b00100, 5},
	{0b100100, 6},
	{0b00101, 5},
	{0b100101, 6},
	{0b100110, 6},
	{0b100111, 6},
	{0b00110, 5},
	{0b1110100, 7},
	{0b1110101, 7},
	{0b101000, 6},
	{0b101001, 6},
	{0b101010, 6},
	{0b00111, 5},
	{0b101011, 6},
	{0b1110110, 7},
	{0b101100, 6},
	{0b01000, 5},
	{0b01001, 5},
	{0b101101, 6},
	{0b1110111, 7},
	{0b1111000, 7},
	{0b1111001, 7},
	{0b1111010, 7},
	{0b1111011, 7},
	{0b111111111111110, 15},
	{0b11111111100, 11},
	{0b11111111111101, 14},
	{0b1111111111101, 13},
	{0b1111111111111111111111111100, 28},
	{0b11111111111111100110, 20},
	{0b1111111111111111010010, 22},
	{0b11111111111111100111, 20},
	{0b11111111111111101000, 20},
	{0b1111111111111111010011, 22},
	{0b1111111111111111010100, 22},
	{0b1111111111111111010101, 22},
	{0b11111111111111111011001, 23},
	{0b1111111111111111010110, 22},
	{0b11111111111111111011010, 23},
	{0b11111111111111111011011, 23},
	{0b11111111111111111011100, 23},
	{0b11111111111111111011101, 23},
	{0b11111111111111111011110, 23},
	{0b111111111111111111101011, 24},
	{0b11111111111111111011111, 23},
	{0b111111111111111111101100, 24},
	{0b111111111111111111101101, 24},
	{0b1111111111111111010111, 22},
	{0b11111111111111111100000, 23},
	{0b111111111111111111101110, 24},
	{0b11111111111111111100001, 23},
	{0b11111111111111111100010, 23},
	{0b11111111111111111100011, 23},
	{0b11111111111111111100100, 23},
	{0b111111111111111011100, 21},
	{0b1111111111111111011000, 22},
	{0b11111111111111111100101, 23},
	{0b1111111111111111011001, 22},
	{0b11111111111111111100110, 23},
	{0b11111111111111111100111, 23},
	{0b111111111111111111101111, 24},
	{0b1111111111111111011010, 22},
	{0b111111111111111011101, 21},
	{0b11111111111111101001, 20},
	{0b1111111111111111011011, 22},
	{0b1111111111111111011100, 22},
	{0b11111111111111111101000, 23},
	{0b11111111111111111101001, 23},
	{0b111111111111111011110, 21},
	{0b11111111111111111101010, 23},
	{0b1111111111111111011101, 22},
	{0b1111111111111111011110, 22},
	{0b111111111111111111110000, 24},
	{0b111111111111111011111, 21},
	{0b1111111111111111011111, 22},
	{0b11111111111111111101011, 23},
	{0b11111111111111111101100, 23},
	{0b111111111111111100000, 21},
	{0b111111111111111100001, 21},
	{0b1111111111111111100000, 22},
	{0b111111111111111100010, 21},
	{0b11111111111111111101101, 23},
	{0b1111111111111111100001, 22},
	{0b11111111111111111101110, 23},
	{0b11111111111111111101111, 23},
	{0b11111111111111101010, 20},
	{0b1111111111111111100010, 22},
	{0b1111111111111111100011, 22},
	{0b1111111111111111100100, 22},
	{0b11111111111111111110000, 23},
	{0b1111111111111111100101, 22},
	{0b1111111111111111100110, 22},
	{0b11111111111111111110001, 23},
	{0b11111111111111111111100000, 26},
	{0b11111111111111111111100001, 26},
	{0b11111111111111101011, 20},
	{0b1111111111111110001, 19},
	{0b1111111111111111100111, 22},
	{0b11111111111111111110010, 23},
	{0b1111111111111111101000, 22},
	{0b1111111111111111111101100, 25},
	{0b11111111111111111111100010, 26},
	{0b11111111111111111111100011, 26},
	{0b11111111111111111111100100, 26},
	{0b111111111111111111111011110, 27},
	{0b111111111111111111111011111, 27},
	{0b11111111111111111111100101, 26},
	{0b111111111111111111110001, 24},
	{0b1111111111111111111101101, 25},
	{0b1111111111111110010, 19},
	{0b111111111111111100011, 21},
	{0b11111111111111111111100110, 26},
	{0b111111111111111111111100000, 27},
	{0b111111111111111111111100001, 27},
	{0b11111111111111111111100111, 26},
	{0b111111111111111111111100010, 27},
	{0b111111111111111111110010, 24},
	{0b111111111111111100100, 21},
	{0b111111111111111100101, 21},
	{0b11111111111111111111101000, 26},
	{0b11111111111111111111101001, 26},
	{0b1111111111111111111111111101, 28},
	{0b111111111111111111111100011, 27},
	{0b111111111111111111111100100, 27},
	{0b111111111111111111111100101, 27},
	{0b11111111111111101100, 20},
	{0b111111111111111111110011, 24},
	{0b11111111111111101101, 20},
	{0b111111111111111100110, 21},
	{0b1111111111111111101001, 22},
	{0b111111111111111100111, 21},
	{0b111111111111111101000, 21},
	{0b11111111111111111110011, 23},
	{0b1111111111111111101010, 22},
	{0b1111111111111111101011, 22},
	{0b1111111111111111111101110, 25},
	{0b1111111111111111111101111, 25},
	{0b111111111111111111110100, 24},
	{0b111111111111111111110101, 24},
	{0b11111111111111111111101010, 26},
	{0b11111111111111111110100, 23},
	{0b11111111111111111111101011, 26},
	{0b111111111111111111111100110, 27},
	{0b11111111111111111111101100, 26},
	{0b11111111111111111111101101, 26},
	{0b111111111111111111111100111, 27},
	{0b111111111111111111111101000, 27},
	{0b111111111111111111111101001, 27},
	{0b111111111111111111111101010, 27},
	{0b111111111111111111111101011, 27},
	{0b1111111111111111111111111110, 28},
	{0b111111111111111111111101100, 27},
	{0b111111111111111111111101101, 27},
	{0b111111111111111111111101110, 27},
	{0b111111111111111111111101111, 27},
	{0b111111111111111111111110000, 27},
	{0b11111111111111111111101110, 26},
}
