FROM registry.access.redhat.com/ubi8/ubi-minimal

# Labels
LABEL name="aws-vmcreate" \
    maintainer="nikhil.com" \
    vendor="nikhil" \
    version="1.0.0" \
    release="1" \
    summary="This service enables provisioning/de-provisioning of AWS cloud vms." \
    description="This service enables provisioning/de-provisioning AWS cloud vms."

# copy code to the build path
USER root
WORKDIR /opt
RUN chgrp -R 0 /opt && \
    chmod -R g=u /opt && \
    chmod +x -R /opt
USER 1001
ENV ec2_tag_key "Name"
ENV ec2_tag_value "nikhil-test"
ENV ec2_command "create"
# ENV ec2_image_id "ami-0d0ca2066b861631c"
# ENV ec2_instance_type "t2.micro"
#can be delete also
# COPY go.* ./
COPY aws-vmcreate .
# ADD data data/
# RUN go mod download
# RUN go build -o aws-vmcreate
CMD ["bash","-c","/opt/aws-vmcreate -c  $ec2_command -n $ec2_tag_key -v $ec2_tag_value"]
