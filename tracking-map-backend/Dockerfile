
FROM golang:1.25-alpine

#Tạo một thư mục tên là /app bên trong Docker để chứa code
WORKDIR /app

#Copy 2 file quản lý thư viện v
COPY go.mod go.sum ./

#tự động tải toàn bộ thư viện 
RUN go mod download

#Copy source code vào trong Docker
COPY . .

#Đóng gói code thành một file là "main"
RUN go build -o main .

#Mở cổng 8080 và chạy file "main"
EXPOSE 8080
CMD ["./main"]