import {
  Collection,
  Filter,
  InsertOneResult,
  MongoClient,
  MongoNetworkError,
  OptionalUnlessRequiredId,
  ServerApiVersion,
  UpdateResult,
  WithId,
} from "mongodb";

const MONGODB_URI = process.env.MONGODB_URI || "";
const AUTH = "auth";

export class Repository<T extends Object> {
  private readonly client: MongoClient;
  private readonly collection: Collection<T>;
  constructor() {
    this.client = new MongoClient(MONGODB_URI, {
      serverApi: {
        version: ServerApiVersion.v1,
      },
    });
    this.client.connect();

    this.client.on("connectionReady", () => {
      console.log("[RepoClient] Connection OK");
    });

    this.client.on("error", (err) => {
      console.error("[RepoClient] ERROR: " + err);
      throw new MongoNetworkError(err.message);
    });

    this.collection = this.client.db().collection<T>(AUTH);
  }

  async getOne(filter: Partial<WithId<T>>): Promise<WithId<T> | null> {
    return this.collection.findOne(filter as Filter<T>);
  }

  async getMany(filters: Partial<WithId<T>>): Promise<WithId<T>[]> {
    return this.collection.find(filters as Filter<T>).toArray();
  }

  async create(requestDto: T): Promise<InsertOneResult<T>> {
    return this.collection.insertOne(requestDto as OptionalUnlessRequiredId<T>);
  }

  async update(_id: string, requestDto: Partial<T>): Promise<UpdateResult<T>> {
    return this.collection.updateOne({ _id } as Filter<T>, requestDto);
  }
}
