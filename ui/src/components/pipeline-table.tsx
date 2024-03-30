import PlayCircleIcon from '@mui/icons-material/PlayCircle';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import { pipeline } from '../models/pipeline';
import { useDockerDesktopClient } from '../common/docker-client';
import { useMutation, useQuery } from '@tanstack/react-query';

interface pipelineRunConfig {
  file: string;
  path: string;
  config: runnerConfig;
}

interface runnerConfig {
  workspace_base: string;
  workspace_path: string;
  secrets: {};
  env: string[];
  privileged: string[];
}

export function PipelineTable() {
  const ddClient = useDockerDesktopClient();

  const getPipelines = async(): Promise<pipeline[]> => {
    try {
      const result = await ddClient.extension.vm?.service?.get("/api/pipeline"); 
      return Promise.resolve(result as pipeline[]);
    } catch (error) {
      return Promise.reject(error)
    }
  }

  const { data } = useQuery({queryKey: ['pipelines'], queryFn: getPipelines})

  const startPipeline = (conf: pipelineRunConfig): Promise<any> =>{
    return ddClient.extension.vm?.service?.post("/api/pipeline/start", conf) as Promise<any>;
  }

  const startPipelineMutation = useMutation({
    mutationFn: startPipeline, 
    onSuccess: () => {
      console.log("pipeline started")
    },
    onError: (error) => {
      console.error("failed to start pipeline:")
      console.error(error)
      ddClient.desktopUI.toast.error(`failed to start pipeline: ${error}`)
    }
  })

  const copyRepo = (conf: {repoPath: string, basePath: string}): Promise<any> => {
    return ddClient.docker.cli.exec("cp", [
      conf.repoPath,
      "mjehanno_woodpecker-ci-desktop-extension-service:"+conf.basePath
    ])
  }

  const copyRepoMutation = useMutation({
    mutationFn: copyRepo,
    onSuccess: () => {
      console.log("repo copied successfully in container")
    }
    })

  const handleStartPipeline = async (pip: pipeline) => {
      console.group("handlestartPipeline")
      const repoPath = pip.path.replace(pip.name, "");
      const arrPath =  repoPath.split("/")
      const repoName = arrPath[arrPath.length -2]
      const basePath = "/home/woody/repos/"

      console.dir({ repoPath, repoName, basePath})

      console.log("copying repo inside container")
      //copy repo to container
      copyRepoMutation.mutate({repoPath, basePath})

      const postData = {
        file: basePath + repoName  +"/woodpecker.yaml",
        path: basePath + repoName ,
        config:{
          workspace_base: basePath + repoName,
          workspace_path: repoName, 
          secrets: {},
          env: [],
          privileged: [
            "plugins/docker", "plugins/gcr", "plugins/ecr", "woodpeckerci/plugin-docker-buildx", "codeberg.org/woodpecker-plugins/docker-buildx"
          ]
        }
      }

      console.log("triggering the pipeline")
      startPipelineMutation.mutate(postData);
      console.groupEnd();
  }

    return (
        <>
  <TableContainer component={Paper}>
      <Table sx={{ minWidth: 650 }} aria-label="simple table">
        <TableHead>
          <TableRow>
            <TableCell>Name</TableCell>
            <TableCell>Status</TableCell>
            <TableCell>Action</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {data?.map((row) => (
            <TableRow
              key={row.id}
              sx={{ '&:last-child td, &:last-child th': { border: 0 } }}
            >
              <TableCell>{row.path}</TableCell>
              <TableCell>{row.status}</TableCell>
              <TableCell>
                <PlayCircleIcon style={{cursor: "pointer"}} onClick={() => handleStartPipeline(row)} />
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
        </>
    )
}