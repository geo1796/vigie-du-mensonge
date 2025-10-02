FROM node:20-alpine
WORKDIR /app
RUN npm install -g @redocly/cli
COPY . ./
ENTRYPOINT ["sh", "./docs.sh", "/app/docs/docs.html"]