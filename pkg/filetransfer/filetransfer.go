package filetransfer

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"unicode/utf8"
)

type FileTransfer struct {
	Port     int
	FileName string
	Listen   bool
	Send     bool
	Hostname string
}

func (ft *FileTransfer) FileTransferInput(fs *flag.FlagSet) {
	fs.IntVar(&ft.Port, "port", 8080, "Provide a port to bind to on this host")
	fs.StringVar(&ft.FileName, "filename", "tempfile1", "Provide the filename you want to transfer/receive.")
	fs.BoolVar(&ft.Listen, "listen", false, "Listen for a file to be transferred to me.")
	fs.BoolVar(&ft.Send, "send", false, "Initiate a connection out and send a file.")
	fs.StringVar(&ft.Hostname, "hostname", "127.0.0.1", "Hostname we are transfering to. IP Address works as well.")
}

func FileTransferLogic(fti *FileTransfer) {
	var (
		bindAddress   = fmt.Sprintf(":%d", fti.Port)
		fileName      = fti.FileName
		senderAddress = fmt.Sprintf("%s:%d", fti.Hostname, fti.Port)
	)

	if fti.Listen && fti.Send {
		log.Fatalln("[*] Either Send or Listen, unable to do both.")
	}

	if fti.Listen {
		ListenerLogic(bindAddress, fileName)
	}

	if fti.Send {
		SenderLogic(senderAddress, fileName)
	}
}

func SenderLogic(senderAddress string, localFileName string) {
	conn, err := net.Dial("tcp", senderAddress)
	if err != nil {
		log.Fatalf("Error connecting: %s\n", err)
	}
	defer conn.Close()

	inputFile, err := os.Open(localFileName)
	if err != nil {
		log.Fatalf("Error unable to open file: %s\n", err)
	}
	defer inputFile.Close()

	// Copy the file content to the TCP connection
	_, err = io.Copy(conn, inputFile)
	if err != nil {
		log.Fatalf("Error sending file: %s\n", err)
	}

	log.Println("File sent successfully!")
}

func ListenerLogic(bindAddress string, fileName string) {
	listener, err := net.Listen("tcp", bindAddress)
	if err != nil {
		log.Fatalf("[-] Error listening: %s", err)
	}
	defer listener.Close()

	log.Printf("[*] Listening for connections on %s", bindAddress)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("[-] Unable to accept connection from %s: %s", conn.RemoteAddr().String(), err)
		}
		go processListenerConnect(conn, fileName)
	}
}

func processListenerConnect(conn net.Conn, fileName string) {
	defer conn.Close()

	log.Printf("Client %s connected. Receiving file...", conn.RemoteAddr())

	// Create a new file to write the received content
	outputFile, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("Error creating file: %s\n", err)
	}
	defer outputFile.Close()

	// Create a buffer reader for the TCP connection
	reader := bufio.NewReader(conn)

	// Create a buffer writer for writing to the file
	writer := bufio.NewWriter(outputFile)

	// Read data from the TCP connection and write to the file
	_, err = io.Copy(writer, reader)
	if err != nil {
		log.Fatalf("Error receiving file: %s\n", err)
	}

	// Flush any unwritten data from the writer to the file
	err = writer.Flush()
	if err != nil {
		log.Fatalf("Error flushing data to file: %s", err)
	}

	log.Println("File received successfully!")

}

// Below should work

func isUTF8(data []byte) bool {
	return utf8.Valid(data)
}

func CheckForBinary(fileInputName string) {
	// Open the input file
	inputFile, err := os.Open(fileInputName)
	if err != nil {
		log.Fatalf("Error unable to open %s: %s", fileInputName, err)
	}
	defer inputFile.Close()

	// Create a buffer reader for the input file
	reader := bufio.NewReader(inputFile)

	// Flag to track if the file is binary
	isBinary := false

	for {
		// Read a chunk of data from the input file
		buffer := make([]byte, 512) // Maximum of 512 bytes for initial check
		bytesRead, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			log.Fatalf("Error unable to read the first 512 bytes: %s", err)
		}

		// Check if the content is valid UTF-8
		if !isUTF8(buffer[:bytesRead]) {
			isBinary = true
			break
		}

		// If the end of the input file is reached, break the loop
		if err == io.EOF {
			break
		}
	}

	if isBinary {
		// log.Println("The file is binary.")
		WriteBinaryFile(fileInputName)

	} else {
		// log.Println("The file is plain text.")
		WriteTextFiles(fileInputName)

	}
}

func SendFile(fileInputName string) {
	// Open the input file
	inputFile, err := os.Open(fileInputName)
	if err != nil {
		log.Fatalf("Error: %s\n", err)
	}
	defer inputFile.Close()

	// Connect to the TCP server
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatalf("Error connecting to server: %s\n", err)
	}
	defer conn.Close()

	// Copy the file content to the TCP connection
	_, err = io.Copy(conn, inputFile)
	if err != nil {
		log.Fatalf("Error sending file: %s\n", err)
	}

	fmt.Println("File sent successfully!")
}

func WriteBinaryFile(fileName string) {
	// Open the input file
	inputFile, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Error Unable to open input file: %s", err)
		return
	}
	defer inputFile.Close()

	// Create the output file
	outputFile, err := os.Create("output.pdf")
	if err != nil {
		log.Fatalf("Error unable to create output file: %s", err)
	}
	defer outputFile.Close()

	// Copy the contents of the input file to the output file
	_, err = io.Copy(outputFile, inputFile)
	if err != nil {
		log.Fatalf("Error unable to copy contents of input to output file: %s", err)
	}

	log.Println("File copied successfully!")
}

func WriteTextFiles(fileName string) {
	//Use this function when working with text files

	// Open the input file
	inputFile, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Error opening the input file: %s", err)
	}
	defer inputFile.Close()

	// Create the output file
	outputFile, err := os.Create("output.txt")
	if err != nil {
		log.Fatalf("Error creating output file: %s", err)
	}
	defer outputFile.Close()

	// Create a reader for the input file
	reader := bufio.NewReader(inputFile)

	// Create a writer for the output file
	writer := bufio.NewWriter(outputFile)

	// Create a dynamic buffer to store file contents
	buffer := make([]byte, 1024) // Initial buffer size

	for {
		// Read from the input file into the buffer
		bytesRead, err := reader.Read(buffer)
		if err != nil && err.Error() != "EOF" {
			log.Fatalf("Error reading from input file into buffer: %s", err)
		}

		// Write the buffer content to the output file
		_, err = writer.Write(buffer[:bytesRead])
		if err != nil {
			log.Fatalf("Error writing buffer content to output file: %s", err)
		}

		// If the end of the input file is reached, break the loop
		if bytesRead < len(buffer) {
			break
		}
	}

	// Flush any unwritten data from the writer to the output file
	err = writer.Flush()
	if err != nil {
		log.Fatalf("Error unable to flush unwritten data from writer to output file: %s\n", err)
	}

	log.Println("File copied successfully!")
}
