FROM microsoft/aspnetcore-build AS build
WORKDIR /workspace
COPY . .
RUN dotnet restore
RUN dotnet publish --configuration release --output ./bin/

FROM microsoft/aspnetcore:1.1
ENV ASPNETCORE_URLS http://*:8080
ENV ASPNETCORE_ENVIRONMENT Production
WORKDIR /app
COPY --from=build /workspace/bin/ .
EXPOSE 8080
ENTRYPOINT ["dotnet", "Podsync.dll"]