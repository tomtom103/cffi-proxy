import asyncio

import httpx

async def async_get(client: httpx.AsyncClient) -> httpx.Response:
    response = await client.get("https://httpbin.org/get")
    print(response.headers)
    print(response.json())
    return response

async def async_main() -> None:
    async with httpx.AsyncClient(proxy="http://0.0.0.0:8000", verify=False) as client:
        tasks = []
        for _ in range(50):
            tasks.append(async_get(client))

        _ = await asyncio.gather(*tasks)
    return

def main() -> None:
    with httpx.Client(proxy="http://0.0.0.0:8000", verify=False) as client:
        response = client.get("https://www.investing.com/equities/nice-information-service-co-ltd-scoreboard")
        print(response.content)
    #     response = client.get("https://httpbin.org/get")
    #     print(response.headers)
    #     print(response.json())

    # asyncio.run(async_main())

if __name__ == "__main__":
    main()
