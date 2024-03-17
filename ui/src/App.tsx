import React from 'react';
import { styled } from '@mui/material/styles';
import Button from '@mui/material/Button';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import { createDockerDesktopClient } from '@docker/extension-api-client';
import { Stack, TextField, Typography } from '@mui/material';
import { useForm } from "react-hook-form";

const client = createDockerDesktopClient();

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

function useDockerDesktopClient() {
  return client;
}

export function App() {
  const { setValue, handleSubmit } = useForm();
  const [response, setResponse] = React.useState<string>();
  const ddClient = useDockerDesktopClient();

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

    const postData = {
      name: data.file.name,
      path: data.file.path,
      type: data.file.type,
    }
    console.log(data.file.name)
    console.log(data.file.path)
    console.log(data.file.type)
    console.log(postData)
    try {
      console.log("trying to upload file")
      const result = await ddClient.extension.vm?.service?.post("/api/upload", postData)
      setResponse(JSON.stringify(result))
      // display docker toast
      // recall get pipeline endpoint (should also be called onMounted)
    } catch (error) {
      console.log("catched error")
      console.error(error)
      setResponse(JSON.stringify(error))
    }
    console.groupEnd()
  }

  const fetchAndDisplayResponse = async () => {
    const result = await ddClient.extension.vm?.service?.get('/hello');
    setResponse(JSON.stringify(result));
  };

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
        <Button variant="contained" onClick={fetchAndDisplayResponse}>
          Call backend
        </Button>

        <TextField
          label="Backend response"
          sx={{ width: 480 }}
          disabled
          multiline
          variant="outlined"
          minRows={5}
          value={response ?? ''}
        />
      </Stack>
    </>
  );
}
