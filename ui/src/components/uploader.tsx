import { styled } from '@mui/material/styles';
import Button from '@mui/material/Button';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useForm } from 'react-hook-form';
import { useDockerDesktopClient } from '../common/docker-client';


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

interface fileContent {
    name: string;
    path: string;
    type: string;
    content: string;
}

export function Uploader() {
    const ddClient = useDockerDesktopClient();
    const queryClient = useQueryClient();
    const { setValue, handleSubmit } = useForm();

    const uploadPipeline = (data: fileContent): Promise<any> =>{
        return ddClient.extension.vm?.service?.post("/api/upload", data) as Promise<any>
    }

    const uploadFileMutation = useMutation({mutationFn: uploadPipeline,
        onSuccess: (data: fileContent) => {
            ddClient.desktopUI.toast.success(`${data.name} file uploaded`);
            queryClient.invalidateQueries({queryKey: ['pipelines']})
        },
        onError: (error, data: fileContent) => {
            console.error(error)
            ddClient.desktopUI.toast.error(`failed to upload ${data.name}`);
        }
    })

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
        } as fileContent
        console.log(postData)
        uploadFileMutation.mutate(postData)
        console.groupEnd()
    }

    return (
        <>
        <Button component="label" role={undefined} variant="contained" tabIndex={-1} startIcon={<CloudUploadIcon />}>
            Upload file
            <VisuallyHiddenInput type="file" onChange={ (e) => {handleChange(e); e.target.value=""}}/>
        </Button>
        </>
    )
}