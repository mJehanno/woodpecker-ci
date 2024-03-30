import { Stack, Typography } from '@mui/material';
import { PipelineTable } from './components/pipeline-table';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Uploader } from './components/uploader';

const queryClient = new QueryClient();

export function App() {

  return (
    <>
      <QueryClientProvider client={queryClient}>
        <Typography variant="h3">Woodpecker pipelines</Typography>
        <Typography variant="body1" color="text.secondary" sx={{ mt: 2 }}>
          Run Woodpecker-CI pipelines from docker desktop.
        </Typography>
        <Stack direction="row" alignItems="start" spacing={2} sx={{ mt: 4 }}>
            <Uploader />
        </Stack>
        <PipelineTable />
      </QueryClientProvider>
    </>
  );
}

