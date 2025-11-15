import { Input, Button, Grid, GridItem, Stack } from "@chakra-ui/react";
import { useState } from "react";
import { useForm, type SubmitHandler } from "react-hook-form";

export default function App() {
  return (
    <Stack w="50%" maxW="540px" marginX="auto" marginTop="20vh" align="center">
      <h1>ushort</h1>
      <Processor />
    </Stack>
  );
}

enum Action {
  QrGenerator,
  UrlShortener,
}

const Processor = () => {
  const { handleSubmit, url, onChangeUrl } = useProcess();

  return (
    <Stack as="form">
      <Input
        placeholder="ushort.com"
        variant="subtle"
        value={url}
        onChange={onChangeUrl}
      />

      <Grid templateColumns={"1fr 1fr"} gap={2}>
        <GridItem>
          <Button
            w="full"
            variant="solid"
            colorPalette={"blue"}
            onClick={handleSubmit(Action.QrGenerator)}
          >
            QR Code
          </Button>
        </GridItem>
        <GridItem>
          <Button
            w="full"
            variant="subtle"
            colorPalette={"blue"}
            onClick={handleSubmit(Action.UrlShortener)}
          >
            Shorten a link
          </Button>
        </GridItem>
      </Grid>
    </Stack>
  );
};

function useProcess() {
  const [url, setUrl] = useState<string>("");
  const onChangeUrl = (e: React.ChangeEvent<HTMLInputElement>) => {
    setUrl(e.target.value);
  };

  const submitUrl = (e: React.FormEvent) => {
    e.preventDefault();
    console.log("submit url:", url);
    fetch("/api/newurl").then(console.log).catch(console.log);
  };

  const submitQR = (e: React.FormEvent) => {
    e.preventDefault();
    console.log("submit QR:", url);
    fetch("/api/qrcode").then(console.log).catch(console.log);
  };

  const handleSubmit = (action: Action) => {
    switch (action) {
      case Action.QrGenerator:
        return submitQR;
      case Action.UrlShortener:
        return submitUrl;
    }
  };

  return { handleSubmit, url, onChangeUrl };
}
