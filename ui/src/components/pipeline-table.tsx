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


export function PipelineTable({pipelines}: {pipelines: pipeline[]}) {
    console.log("pipeline-table constructor")
    const ddClient = useDockerDesktopClient();
    const handleStartPipeline = async (pip: pipeline) => {
        console.group("handlestartPipeline")
        const repoPath = pip.path.replace(pip.name, "");
        const arrPath =  repoPath.split("/")
        const repoName = arrPath[arrPath.length -2]
        const basePath = "/home/woody/repos/"

        console.dir({ repoPath, repoName, basePath})

        console.log("copying repo inside container")
        //copy repo to container
        await ddClient.docker.cli.exec("cp", [
          repoPath,
          "mjehanno_woodpecker-ci-desktop-extension-service:"+basePath
        ])

        const postData = {
          file: basePath + repoName  +"/woodpecker.yaml",
          path: basePath + repoName ,
          config:{
            workspace_base: basePath + repoName,
            workspace_path: repoName, 
            secrets: {},
            env: [],
            priviliged: [
              "plugins/docker", "plugins/gcr", "plugins/ecr", "woodpeckerci/plugin-docker-buildx", "codeberg.org/woodpecker-plugins/docker-buildx"
            ]
          }
        }

        console.log("triggering the pipeline")
        //start pipeline
        try {
          await ddClient.extension.vm?.service?.post("/api/pipeline/start", postData);
          console.log("pipeline started")
        }catch(error) {
          console.error("failed to start pipeline:")
          console.error(error)
          ddClient.desktopUI.toast.error(`failed to start pipeline: ${error}`)
        }

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
          {pipelines.map((row) => (
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