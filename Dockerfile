From ibmterraform/terraform-provider-ibm-docker:latest


ENV API_REPO /go/src/github.com/terraform-provider-ibm-api
COPY . $API_REPO
RUN cd $API_REPO && \
    go build -o apiserver
EXPOSE 9080
WORKDIR $API_REPO
CMD ["./apiserver"]