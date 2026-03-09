
FROM node:18-alpine

#Tạo thư mục làm việc
WORKDIR /app

#Copy file danh sách thư viện 
COPY package*.json ./

#Cài đặt thư viện
RUN npm install

#Copy code vào trong Docker
COPY . .

#Mở cổng 3000 
EXPOSE 3000

#Chạy server 
CMD ["npm", "start"] 
