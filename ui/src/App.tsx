import {
  Input,
  Button,
  Grid,
  GridItem,
  Stack,
  Link,
  Text,
} from "@chakra-ui/react";
import { useState } from "react";

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

type ActionResult = {
  action: Action;
  data: string;
};

const Processor = () => {
  const { handleSubmit, url, onChangeUrl, result, loading } = useProcess();

  return (
    <>
      <Stack as="form">
        <Input
          placeholder="ushort.com"
          variant="subtle"
          value={url}
          onChange={onChangeUrl}
          disabled={loading}
        />

        <Grid templateColumns={"1fr 1fr"} gap={2}>
          <GridItem>
            <Button
              w="full"
              variant="solid"
              colorPalette={"blue"}
              onClick={handleSubmit(Action.QrGenerator)}
              disabled={loading}
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
              disabled={loading}
            >
              Shorten a link
            </Button>
          </GridItem>
        </Grid>
      </Stack>

      {result && <ResultDisplay {...result} />}
    </>
  );
};

function ResultDisplay({ action, data }: ActionResult) {
  switch (action) {
    case Action.QrGenerator:
      return <>fooooo</>;
    case Action.UrlShortener:
      const url = `http://localhost:5173/${data}`;
      return (
        <Stack align="center" mt={3}>
          <Text>Shortened URL:</Text>
          <Link href={url} target="_blank">
            {url}
          </Link>
        </Stack>
      );
  }
}

function useProcess() {
  const [url, setUrl] = useState<string>("");
  const onChangeUrl = (e: React.ChangeEvent<HTMLInputElement>) => {
    setUrl(e.target.value);
  };

  const [result, setResult] = useState<ActionResult | undefined>();
  const [loading, setLoading] = useState<boolean>(false);
  const submitUrl = async (e: React.FormEvent) => {
    e.preventDefault();
    const postUrl = `/api/shorturl?url=${encodeURI(url)}`;

    setLoading(true);
    try {
      const response = await fetch(postUrl, { method: "POST" });
      const hash = await response.text();
      setResult({ action: Action.UrlShortener, data: hash });
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
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

  return { handleSubmit, url, onChangeUrl, result, loading };
}
