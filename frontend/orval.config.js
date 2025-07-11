module.exports = {
  backend: {
    input: '../docs/swagger.yaml',
    output: {
      mode: 'tags',
      target: './lib/api/orval/',
      client: 'axios',
      mock: false,
      prettier: true,
      override: {
        mutator: {
          path: './lib/api/custom-axios.ts',
          name: 'customAxios',
        },
      },
    },
  },
};