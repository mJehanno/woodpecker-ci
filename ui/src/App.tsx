import React, { useEffect } from 'react';
import { styled } from '@mui/material/styles';
import Button from '@mui/material/Button';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import { Stack, Typography } from '@mui/material';
import { useForm } from "react-hook-form";
import { pipeline } from './models/pipeline';
import { PipelineTable } from './components/pipeline-table';
import { useDockerDesktopClient } from './common/docker-client';



const VisuallyHiddenInput = styled('input')({
  clip: 'rect(0 0 0 0)',
  clipPath: 'inset(50%)',
  height: 1,
  overflow: 'hidden',
  position: 'absolute',
  bottom: 0,
  left: 0,
  whiteSpace: 'nowrap',
  width: 1,
});


export function App() {
  const { setValue, handleSubmit } = useForm();
  const ddClient = useDockerDesktopClient();
  const [pipelines, setPipelines] = React.useState<pipeline[]>([]);
  const [upload, setUpload] = React.useState<boolean>(false)

  useEffect(() => {
    getPipelines()
  }, [setValue, setPipelines, upload])

  const getPipelines = async() => {
    console.group("getPipelines")
    ddClient.extension.vm?.service?.get("/api/pipeline").then((x: unknown) => {
    console.log(x)
    setPipelines(x as pipeline[])
    console.groupEnd()
    })
  }

  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    console.group("handleChange")
    console.log(event.target.files);
    if (event.target.files){
      setValue("file", event.target?.files[0]);
    }
    handleSubmit(handleUpload)();
    console.groupEnd()
  }

  

  const handleUpload = async (data:any) => {
    console.group("handleUpload")
    const content = await data.file.text()
    const postData = {
      name: data.file.name,
      path: data.file.path,
      type: data.file.type,
      content: content 
    }
    console.log(postData)

    try {
      const result = await ddClient.extension.vm?.service?.post("/api/upload", postData)
      ddClient.desktopUI.toast.success(`${postData.name} file uploaded`);
      // recall get pipeline endpoint (should also be called onMounted)
    } catch (error) {
      console.error(error)
      ddClient.desktopUI.toast.error(`failed to upload ${postData.name}`);
    }
    setUpload(true)
    console.groupEnd()
  }


  return (
    <>
      <Typography variant="h3">Woodpecker pipelines</Typography>
      <Typography variant="body1" color="text.secondary" sx={{ mt: 2 }}>
        Run Woodpecker-CI pipelines from docker desktop.
      </Typography>
      <Stack direction="row" alignItems="start" spacing={2} sx={{ mt: 4 }}>
        <Button component="label" role={undefined} variant="contained" tabIndex={-1} startIcon={<CloudUploadIcon />}>
          Upload file
          <VisuallyHiddenInput type="file" onChange={handleChange}/>
        </Button>
      </Stack>
      <PipelineTable pipelines={pipelines} />
    </>
  );
}

