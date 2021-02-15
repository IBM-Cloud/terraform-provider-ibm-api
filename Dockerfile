FROM ibmterraform/terraform-provider-ibm-docker:latest

# We can keep these. Will be useful 
ARG gitSHA
ARG travisBuildNo
ARG buildDate

ARG API_REPO=/go/src/github.com/terraform-provider-ibm-api

COPY . $API_REPO

ENV GOPATH="/go"
ENV PATH=$PATH:/usr/local/go/bin:/go/bin

# Compiles and installs the packages named by the import paths,
# along with their dependencies.
# -o apiserver to change the name .. change in Makefile as well then. 
RUN cd $API_REPO && \
    go install -v -ldflags "-X main.commit=${gitSHA} -X main.travisBuildNumber=${travisBuildNo} -X main.buildDate=${buildDate}"

EXPOSE 9080

# RUN chmod -R 775 /go
# RUN addgroup -g 1001 -S appuser && adduser -u 1001 -S appuser -G appuser
# RUN chown -R appuser:appuser /go
# USER appuser

WORKDIR $API_REPO

# Run the application as root process
# To pass HTTP addr and port other than localhost:8080, Docker run with flags
CMD ["terraform-provider-ibm-api"]