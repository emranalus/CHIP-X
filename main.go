package main

import (
	"encoding/binary"
	"fmt"
	"image/color"
	"io"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/hajimehoshi/ebiten/v2"
)

// opcodes are 16-bit
// I is the index register PC is programme counter each is 16-bits
// stack has 16 levels because there is a stack there needs to be a stack pointer
// memory addresses are 8-bits and there is 4096 of them
// registers are 8-bits and there is 16 of them 15 are general pupose last one VF is for carry flags
// Graphics are black and white an screen is 64x32 pixels
var (
    // Hardware Parameters
    Opcode                  uint16
    I                       uint16
    PC                      uint16
    Memory[4096]            uint8
    V[16]                   uint8
    Screen[64][32]         uint8
    DelayTimer              uint8
    SoundTimer              uint8
    Key[16]                 uint8

    // Emulator Parameters
    ReturnAddr              uint16
    ByteIndex               uint8
    C8Width                 int
    C8Heigth                int
    OpcodeFetcher[2]        byte
    CanFetch                bool
    FontArray[75]           uint8
)

// Hardware Memory Map Cheatsheat
// 0x000-0x1FF - Chip 8 interpreter (contains font set in emulator)
// 0x050-0x0A0 - Used for the built in 4x5 pixel font set (0-F)
// 0x200-0xE8F - Program ROM and work RAM

const (
    programmeMemoryBegins   uint16 = 0x200
    programmeMemoryEnds     uint16 = 0xE8F
    programmeMemoryLength   uint16 = programmeMemoryEnds - programmeMemoryBegins 
)

const (
    KeyA ebiten.Key = 0
    KeyS ebiten.Key = 18
    KeyD ebiten.Key = 3
    KeyF ebiten.Key = 5
    KeyQ ebiten.Key = 16
    KeyW ebiten.Key = 22
    KeyE ebiten.Key = 4
    KeyR ebiten.Key = 17
    KeyZ ebiten.Key = 25
    KeyX ebiten.Key = 23
    KeyC ebiten.Key = 2
    KeyV ebiten.Key = 21
    Key1 ebiten.Key = 44
    Key2 ebiten.Key = 45
    Key3 ebiten.Key = 46
    Key4 ebiten.Key = 47
)

//////////////// Game Engine Object ////////////////
type Game struct{}

func LoadFont()  {
   
    // Load font into emulator 
    // 0x0
    FontArray[0] = 0b11100000
    FontArray[1] = 0b10100000
    FontArray[2] = 0b10100000
    FontArray[3] = 0b10100000
    FontArray[4] = 0b11100000
    // 0x1
    FontArray[5] = 0b00100000
    FontArray[6] = 0b00100000
    FontArray[7] = 0b00100000
    FontArray[8] = 0b00100000
    FontArray[9] = 0b00100000
    // 0x2
    FontArray[10] = 0b11100000
    FontArray[11] = 0b00100000
    FontArray[12] = 0b11100000
    FontArray[13] = 0b10000000
    FontArray[14] = 0b11100000
    // 0x3
    FontArray[15] = 0b11100000
    FontArray[16] = 0b00100000
    FontArray[17] = 0b11100000
    FontArray[18] = 0b00100000
    FontArray[19] = 0b11100000
    // 0x4
    FontArray[20] = 0b10100000
    FontArray[21] = 0b10100000
    FontArray[22] = 0b11100000
    FontArray[23] = 0b00100000
    FontArray[24] = 0b00100000
    // 0x5
    FontArray[25] = 0b11100000
    FontArray[26] = 0b10000000
    FontArray[27] = 0b11100000
    FontArray[28] = 0b00100000
    FontArray[29] = 0b11100000
    // 0x6
    FontArray[30] = 0b11100000
    FontArray[31] = 0b10000000
    FontArray[32] = 0b11100000
    FontArray[33] = 0b10100000
    FontArray[34] = 0b11100000
    // 0x7
    FontArray[35] = 0b11100000
    FontArray[36] = 0b00100000
    FontArray[37] = 0b00100000
    FontArray[38] = 0b00100000
    FontArray[39] = 0b00100000
    // 0x8
    FontArray[40] = 0b11100000
    FontArray[41] = 0b10100000
    FontArray[42] = 0b11100000
    FontArray[43] = 0b10100000
    FontArray[44] = 0b11100000
    // 0x9
    FontArray[45] = 0b11100000
    FontArray[46] = 0b10100000
    FontArray[47] = 0b11100000
    FontArray[48] = 0b00100000
    FontArray[49] = 0b00100000
    // 0xA
    FontArray[50] = 0b11100000
    FontArray[51] = 0b10100000
    FontArray[52] = 0b11100000
    FontArray[53] = 0b10100000
    FontArray[54] = 0b10100000
    // 0xB
    FontArray[55] = 0b10000000
    FontArray[56] = 0b10000000
    FontArray[57] = 0b11100000
    FontArray[58] = 0b10100000
    FontArray[59] = 0b11100000
    // 0xC
    FontArray[60] = 0b11100000
    FontArray[61] = 0b10000000
    FontArray[62] = 0b10000000
    FontArray[63] = 0b10000000
    FontArray[64] = 0b11100000
    // 0xD
    FontArray[65] = 0b00100000
    FontArray[66] = 0b00100000
    FontArray[67] = 0b11100000
    FontArray[68] = 0b10100000
    FontArray[69] = 0b11100000
    // 0xF
    FontArray[70] = 0b11100000
    FontArray[71] = 0b10000000
    FontArray[72] = 0b11100000
    FontArray[73] = 0b10000000
    FontArray[74] = 0b10000000

    // Load font into device memory
    for i:=0;i<len(FontArray);i++ {
        Memory[i] = FontArray[i]
    } 

}

func init()  {

    // Load programme into device memory
    file, err := os.Open("1-chip8-logo.ch8")
    if err != nil {
        fmt.Println(err)
    }

    bytes, err := io.ReadAll(file)
    if err != nil {
        fmt.Println(err)
    }

    for i:=0;i<len(bytes);i++ {
        Memory[int(programmeMemoryBegins) + i] = bytes[i]
    }

    //fmt.Println(Memory[programmeMemoryBegins:programmeMemoryEnds])
    //os.Exit(0)

    // Initialize variables
    InitializeSystemVariables()

    // Manually load test programme to memory
   // Memory[0] = 255
   // I = 0
   // 
   // ins := IntToByteArray(MOV)
   // Memory[PC] = ins[0]
   // Memory[PC+1] = ins[1]

   // ins = IntToByteArray(EXIT)
   // Memory[PC+4] = ins[0]
   // Memory[PC+5] = ins[1]

   // jmp, _ := strconv.Atoi("fff")
   // Memory[jmp] = ins[0]
   // Memory[jmp+1] = ins[1]

}

// Emulate the clock cycle
// Ebitengine caps Update() to 60 FPS by default
func (g *Game) Update() error {

    if CanFetch {
        FetchOpcode()
    }

    ExecuteOpcode()

    if DelayTimer > 0 {
        DelayTimer--
    }

    if SoundTimer > 0 {
        SoundTimer--
    }

    //time.Sleep(time.Second)

    return nil    
}

func (g *Game) Draw(screen *ebiten.Image) {
    
    // Scan each row and light up pixels according to the screen array 
    for i:=0; i < C8Width; i++ {

        for j:=0; j < C8Heigth; j++ {

            if Screen[i][j] == 0 {
                // If pixel is 0 set pixel to black
                screen.Set(i, j, color.Black)

            } else {
                // If pixel is not 0 set pixel to white
                screen.Set(i, j, color.White)
                
            }

        }

    }

}

func (g *Game) Layout(outsideWidth, outsideHeigh int) (screenWidth, screenHeight int) {
    return 64, 32
}
//////////////// Game Engine Object ////////////////

func InitializeSystemVariables() {
    
    PC = programmeMemoryBegins
    ByteIndex = 0
    C8Width = 64
    C8Heigth = 32
    CanFetch = true

    Key[0] = uint8(KeyA)
    Key[1] = uint8(KeyS)
    Key[2] = uint8(KeyD)
    Key[3] = uint8(KeyF)
    Key[4] = uint8(KeyQ)
    Key[5] = uint8(KeyW)
    Key[6] = uint8(KeyE)
    Key[7] = uint8(KeyR)
    Key[8] = uint8(KeyZ)
    Key[9] = uint8(KeyX)
    Key[10] = uint8(KeyC)
    Key[11] = uint8(KeyV)
    Key[12] = uint8(Key1)
    Key[13] = uint8(Key2)
    Key[14] = uint8(Key3)
    Key[15] = uint8(Key4)

}

func ExecuteOpcode() {

    ByteCode := HexToString(Opcode)
    //fmt.Println(ByteCode)

    // Opcodes
    // In Golang slicing is right-hand inclusive
    if len(ByteCode) < 4 {
        
        // 00E0 - Clear screen
        if ByteCode == "e0" {

            // Scan each row and set them to zero(will be drawn as black) 
            for i:=0; i < C8Width; i++ {
    
                for j:=0; j < C8Heigth; j++ {
    
                    Screen[i][j] = 0                    
    
                }

        }

        } else if ByteCode == "ee" {
            // 00EE - Return from subroutine
            PC = ReturnAddr

        } else {
            // 0NNN
            jmp := hexToInt(ByteCode[1:])

            PC = uint16(jmp)
            
        }

    }
    
    // 1NNN
    if ByteCode[0] == '1' {
        jmp := hexToInt(ByteCode[1:])

        PC = uint16(jmp)
    }

    // 2NNN
    if ByteCode[0] == '2' {
        ReturnAddr = PC
        jmp := hexToInt(ByteCode[1:])

        PC = uint16(jmp)
    }

    // 3XNN - Skips the next instruction if VX == NN 
    if ByteCode[0] == '3' {
        addr := hexToInt(ByteCode[2:]) 
        index := hexToInt(ByteCode[1:2])

        // Check if VX == NN
        if V[index] == uint8(addr){
            PC += 2     // Because each instruction is 2 bytes long and a single memory adress is 1 byte we increase PC by 2
        }
    }

    // 4XNN - Skips the next instruction if VX != NN
    if ByteCode[0] == '4' {
        addr := hexToInt(ByteCode[2:]) 
        index := hexToInt(ByteCode[1:2])

        // Check if VX != NN
        if V[index] != uint8(addr){
            PC += 2     // Because each instruction is 2 bytes long and a single memory adress is 1 byte we increase PC by 2
        }
    }

    // 5XY0 - Skips the next instruction if VX == VY
    if ByteCode[0] == '5' && ByteCode[3] == '0' {
        vy := hexToInt(ByteCode[2:3]) 
        vx := hexToInt(ByteCode[1:2])

        // Check if VX == VY
        if V[vx] == V[vy]{
            PC += 2     // Because each instruction is 2 bytes long and a single memory adress is 1 byte we increase PC by 2
        }
    }

    // 6XNN - Set VX to NN
    if ByteCode[0] == '6' {
        nn := hexToInt(ByteCode[1:]) 
        vx := hexToInt(ByteCode[1:2])
       
        V[vx] = uint8(nn)
    }

    // 7XNN - Add NN to VX
    if ByteCode[0] == '7' {
        nn := hexToInt(ByteCode[2:]) 
        vx := hexToInt(ByteCode[1:2])

        // If sum exceeds 8-bit its reduced by modulo 256
        if V[vx] + uint8(nn) > 255 {

            reduct := uint16(V[vx] + uint8(nn)) % 256
            V[vx] = 255
            V[vx] -= uint8(reduct)
        
        } else {

            V[vx] += uint8(nn)
        
        }
        
    }

    // 8XY0 - Set VX to the value of VY
    if ByteCode[0] == '8' && ByteCode[3] == '0' {
        vy := hexToInt(ByteCode[2:3]) 
        vx := hexToInt(ByteCode[1:2])
        
        V[vx] = V[vy]
    }

    // 8XY1 - Set VX to (VX |= VY) bitwise OR operation
    if ByteCode[0] == '8' && ByteCode[3] == '1' {
        vy := hexToInt(ByteCode[2:3]) 
        vx := hexToInt(ByteCode[1:2])
        
        V[vx] |= V[vy]
    }

    // 8XY2 - Set VX to (VX &= VY) bitwise AND operation
    if ByteCode[0] == '8' && ByteCode[3] == '2' {
        vy := hexToInt(ByteCode[2:3]) 
        vx := hexToInt(ByteCode[1:2])
        
        V[vx] &= V[vy]
    }

    // 8XY3 - Set VX to (VX ^= VY) bitwise XOR operation
    if ByteCode[0] == '8' && ByteCode[3] == '3' {
        vy := hexToInt(ByteCode[2:3]) 
        vx := hexToInt(ByteCode[1:2])
        
        V[vx] ^= V[vy]
    }

    // 8XY4 - Add VY to VX (VX += VY) set carry flag to 1 if there is 0 if there is not a carry
    if ByteCode[0] == '8' && ByteCode[3] == '4' {
        vy := hexToInt(ByteCode[2:3]) 
        vx := hexToInt(ByteCode[1:2])
       
        // If there is a carry set the flag
        if (V[vx] + V[vy]) > 255 {
            // Addition
            sum := IntegerToBinary(uint(vx + vy))
            intSum, _ := strconv.Atoi(sum[len(sum)-8:])

            V[vx] = uint8(intSum)
            V[15] = 1
        } else {
            V[vx] += V[vy]
            V[15] = 0
        }
    }

    // 8XY5 - Subtract VY from VY set carry to 0 when there is a borrow and 1 if there is not
    if ByteCode[0] == '8' && ByteCode[3] == '5' {
        vy := hexToInt(ByteCode[2:3]) 
        vx := hexToInt(ByteCode[1:2])

        // If there is borrow set VF to 0 else if there is no borrow set VF to 1
        borrow := (^V[vx]) & V[vy]
        if borrow != 0 {
            V[15] = 0
        } else if borrow == 0 {
            V[15] = 1
        }

        // Subtraction
        V[vx] -= V[vy]
    }

    // 8XY6 - Stores the least significant bit of VX in VF and then shifts VX to the right by 1
    if ByteCode[0] == '8' && ByteCode[3] == '6' {
        vx := hexToInt(ByteCode[1:2])

        // Take least significant bit of VX and put it in VF
        b := byte(V[vx])
        bi := string(b)[len(string(b))-1]
        bit, _ := strconv.ParseUint(string(bi), 2, 1)
        V[15] = uint8(bit)

        // Shift VX to the right by 1 bit
        V[vx] >>= 1
    }

    // 8XY7 - Sets VX to VY minus VX. VF is set to 0 when there's a borrow, and 1 when there is not.
    if ByteCode[0] == '8' && ByteCode[3] == '7' {
        vy := hexToInt(ByteCode[2:3]) 
        vx := hexToInt(ByteCode[1:2])

        // If there is borrow set VF to 0 else if there is no borrow set VF to 1
        borrow := (^V[vy]) & V[vx]
        if borrow != 0 {
            V[15] = 0
        } else if borrow == 0 {
            V[15] = 1
        }

        // Subtraction
        V[vx] = V[vy] - V[vx]
    }

    // 8XYE - Stores the most significant bit of VX in VF and then shifts VX to the left by 1
    if ByteCode[0] == '8' && ByteCode[3] == 'e' {
        vx := hexToInt(ByteCode[1:2])
      
        // Take most significant bit of VX and put it in VF
        b := byte(V[vx])
        bi := string(b)[0]
        bit, _ := strconv.ParseUint(string(bi), 2, 1)
        V[15] = uint8(bit)

        // Shift VX to the left by 1 bit
        V[vx] <<= 1
    }

    // 9XY0 - Skips the next instruction if VX does not equal VY.
    if ByteCode[0] == '9' && ByteCode[3] == '0' {
        vy := hexToInt(ByteCode[2:3]) 
        vx := hexToInt(ByteCode[1:2])

        // Check if VX == VY
        if V[vx] != V[vy]{
            PC += 2     // Because each instruction is 2 bytes long and a single memory adress is 1 byte we increase PC by 2
        }
    }

    // ANNN - Sets I to the address NNN
    if ByteCode[0] == 'a' {
        nnn := hexToInt(ByteCode[1:])

        I = uint16(nnn)
    }

    // BNNN - Jumps to the address (NNN + V0)
    if ByteCode[0] == 'b' {
        nnn := hexToInt(ByteCode[1:])

        PC = uint16(V[0]) + uint16(nnn)
    }

    // CXNN - Sets VX to the result of a bitwise and operation on a random number (Typically: 0 to 255) and NN
    if ByteCode[0] == 'c' {
        vx := hexToInt(ByteCode[1:2])
        nn := hexToInt(ByteCode[2:])
       
        // Generate random number
        rand.Seed(time.Now().UnixNano())
        v := rand.Intn(255-0) + 0

        // Set VX
        V[vx] = uint8(v) & uint8(nn)
    }

    // DXYN - Draws a sprite at coordinate (VX, VY) that has a width of 8 pixels and a height of N pixels.
    // Each row of 8 pixels is read as bit-coded starting from memory location I.
    // I value does not change after the execution of this instruction. 
    // As described above, VF is set to 1 if any screen pixels are flipped from set to unset when the sprite is drawn, and to 0 if that does not happen
    if ByteCode[0] == 'd' {
        vx := hexToInt(ByteCode[1:2])
        vy := hexToInt(ByteCode[2:3])
        h := hexToInt(ByteCode[3:])

        //fmt.Println(vx, vy, h)
//        os.Exit(0)

        // Render sprite starting from (vx, vy) each row is 8 pixel wide
        // Row pixels are bit-coded 1 is white 0 is black 
        for i:=0; i < int(h); i++ {
            
            BitString := fmt.Sprintf("%08b", Memory[I+uint16(i)])

            //fmt.Println(BitString)
//           os.Exit(0)
            
            // Render individual row pixel by pixel
            // If there is a different color on the screen VF is set to 1 otherwise its set to 0
            for j:=0; j < len(BitString); j++ {

                if string(BitString[j]) == "0" {

                    if Screen[V[vx]+uint8(j)][V[vy]+uint8(i)] == 1 {
                        V[15] = 1
                    } else {
                        V[15] = 0
                    }

                    Screen[V[vx]+uint8(j)][V[vy]+uint8(i)] = 0

                } else if string(BitString[j]) == "1" {

                    if Screen[V[vx]+uint8(j)][V[vy]+uint8(i)] == 0 {
                        V[15] = 1
                    } else {
                        V[15] = 0
                    }

                    Screen[V[vx]+uint8(j)][V[vy]+uint8(i)] = 1

                }

            }

        }

    }
     
    // EX9E - Skips the next instruction if the key stored in VX is pressed 
    if ByteCode[0] == 'e' && ByteCode[2:] == "9e" {
        vx := hexToInt(ByteCode[1:2])

        // Check the key stored in VX
        if ebiten.IsKeyPressed(ebiten.Key(V[vx])) {
            PC += 2
        }

    }

    // EXA1 - Skips the next instruction if the key stored in VX is not pressed 
    if ByteCode[0] == 'e' && ByteCode[2:] == "a1" {
        vx := hexToInt(ByteCode[1:2])

        // Check the key stored in VX
        if ebiten.IsKeyPressed(ebiten.Key(V[vx])) == false {
            PC += 2
        }

    }

    // FX07 - Sets VX to the value of the delay timer
    if ByteCode[0] == 'f' && ByteCode[2:] == "07" {
        vx := hexToInt(ByteCode[1:2])

        V[vx] = DelayTimer
    }

    // FX0A - A key press is awaited, and then stored in VX 
    if ByteCode[0] == 'f' && ByteCode[2:] == "0a" {
        vx := hexToInt(ByteCode[1:2])

        CanFetch = false

        for i:=0;i<len(Key);i++ {
            
            if ebiten.IsKeyPressed(ebiten.Key(Key[i])) {
                V[vx] = Key[i]
                CanFetch = true
                break
            }

        }

    }

    // FX15 - Sets the delay timer to VX
    if ByteCode[0] == 'f' && ByteCode[2:] == "15" {
        vx := hexToInt(ByteCode[1:2])

        DelayTimer = V[vx]
    }

    // FX18 - Sets the sound timer to VX
    if ByteCode[0] == 'f' && ByteCode[2:] == "18" {
        vx := hexToInt(ByteCode[1:2])

        SoundTimer = V[vx]
    }

    // FX1E - Adds VX to I
    if ByteCode[0] == 'f' && ByteCode[2:] == "1e" {
        vx := hexToInt(ByteCode[1:2])

        I += uint16(V[vx])
    }

    // FX29 - Sets I to the location of the sprite for the character which is in VX
    if ByteCode[0] == 'f' && ByteCode[2:] == "29" {
        vx := hexToInt(ByteCode[1:2])

        I = uint16(V[vx])
    }

    // FX33 - Stores the binary-coded decimal representation of VX
    // with the hundreds digit in memory at location in I 
    // the tens digit at location I+1, and the ones digit at location I+2.
    if ByteCode[0] == 'f' && ByteCode[2:] == "33" {
        vx := hexToInt(ByteCode[1:2])
        str := fmt.Sprintf("%03d", V[vx])

        // Binary coded decimal is same code according to the truth table for every digid stored in the Memory[I+i]
        for i:=0; i < len(str); i++ {
            
            atoi, _ := strconv.Atoi(string(str[i]))
            Memory[I+uint16(i)] = uint8(atoi)

        }

    }

    // FX55 - Stores from V0 to VX (including VX) in memory, starting at address I
    // The offset from I is increased by 1(I+i) for each value written, but I itself is left unmodified.
    if ByteCode[0] == 'f' && ByteCode[2:] == "55" {
        vx := hexToInt(ByteCode[1:2])

        // What dafuq I just did???
        Offset := 0
        for i:=len(V[:vx]); i > 0; i-- {
            Memory[I+uint16(Offset)] = V[vx-i]
            Offset++
        }

    }

    // FX65 - Fills from V0 to VX (including VX) with values from memory, starting at address I 
    // The offset from I is increased by 1 for each value read
    if ByteCode[0] == 'f' && ByteCode[2:] == "65" {
        vx := hexToInt(ByteCode[1:2])

        // Dafuq inverted
        Offset := 0
        for i:=len(V[:vx]); i > 0; i-- {
            V[vx-i] = Memory[I+uint16(Offset)]
            Offset++
        }

    }

}

func IntegerToBinary(n uint) string {
   return strconv.FormatInt(int64(n), 2)
}

func hexToInt(hexStr string) int {
	result, _ := strconv.ParseUint(strings.Replace(hexStr, "0x", "", -1), 16, 64)
	return int(result)
}

func IntToByteArray(num uint16) []byte {
    size := int(unsafe.Sizeof(num))
    arr := make([]byte, size)
    for i := 0 ; i < size ; i++ {
        byt := *(*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&num)) + uintptr(i)))
        arr[i] = byt
    }
    return arr
}

func HexToString(opcode uint16) string {
    return strconv.FormatInt(int64(opcode), 16)
}


func FetchOpcode() {
   
     OpcodeFetcher[ByteIndex] = Memory[PC]
     PC++
 
     OpcodeFetcher[ByteIndex+1] = Memory[PC]
     PC++

     // .ch8 files stores bytes in big endian fashion.
     Opcode = binary.BigEndian.Uint16(OpcodeFetcher[:])

}

func main() {

    // Run Game Engine
    game := &Game{}
    ebiten.SetWindowSize(640, 320)
    ebiten.SetWindowTitle("CHIP-X")
    if err := ebiten.RunGame(game); err != nil {
        log.Fatal(err)
    }

}

