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
import { callLinter } from '../common/pipeline-linter';


export function PipelineTable({pipelines}: {pipelines: pipeline[]}) {
    console.log("pipeline-table constructor")
    const ddClient = useDockerDesktopClient();
    const handleStartPipeline = async (pip: pipeline) => {
        console.group("handlestartPipeline")
        // lint file
        try {
            callLinter(ddClient, pip.path, pip.name);
        } catch(error) {
            return;
        }

        //spanw server
        // spawn agent
        try {
            const output = await ddClient.docker.cli.exec("run", [
                "-e WOODPECKER_SERVER=localhost:9000",
                `-e WOODPECKER_AGENT_SECRET="a_secret_not_so_secret_yet"`,
                "-e WOODPECKER_MAX_WORKFLOWS=4",
                "woodpeckerci/woodpecker-agent"
            ]);
            console.log(output);
            console.error(output.stderr);
        } catch (error) {
            console.log(error);
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
              <TableCell>{row.name}</TableCell>
              <TableCell>{row.status}</TableCell>
              <TableCell>
                <PlayCircleIcon onClick={() => handleStartPipeline(row)} />
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
        </>
    )
}