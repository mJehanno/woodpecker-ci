import { DockerDesktopClient } from "@docker/extension-api-client-types/dist/v1";

export async function callLinter(client :DockerDesktopClient, path:string, name:string) {
    console.group("callLinter");

    const pipelineFile = await getPipelineFileContent(path);

    const postData = {
      name: name,
      path: path,
      content: pipelineFile, 
    };

    try {
      console.log("linting file");
      client.extension.vm?.service?.post("/api/pipeline/lint", postData);
      console.groupEnd();
      return Promise.resolve();
    } catch (error) {
      console.error(error);
      client.desktopUI.toast.error(`failed to lint ${postData.name} : ${error}`);
      console.groupEnd();
      return Promise.reject();
    }
}

async function getPipelineFileContent(path: string): Promise<string>{
    console.group("getPipelineFileContent");
    const fileUrl = "file://" + path;
    console.log(fileUrl);
    const response = await fetch(fileUrl);
    const pipelineFile = await response.text();
    console.groupEnd();
    return Promise.resolve(pipelineFile);
}