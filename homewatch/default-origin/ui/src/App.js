import React from 'react';
import {
  ChakraProvider,
  theme,
} from '@chakra-ui/react';
import Cameras from "./Cameras"
const appVersion = process.env.REACT_APP_VERSION

function App() {
  return (
    <ChakraProvider theme={theme}>
      <h1>Version {appVersion ? appVersion : "LocalVersion"}</h1>
      <Cameras />
    </ChakraProvider>
  );
}

export default App;
