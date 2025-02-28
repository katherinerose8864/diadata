CREATE EXTENSION "pgcrypto";


-- Table asset is the single source of truth for all assets handled at DIA.
-- If a field is not case sensitive (such as address for Ethereum) it should
-- be all lowercase for consistency reasons.
-- Otherwise it must be as defined in the underlying contract.
CREATE TABLE asset (
    asset_id UUID DEFAULT gen_random_uuid(),
    symbol text not null,
    name text not null,
    decimals text,
    blockchain text,
    address text not null,
    UNIQUE (asset_id),
    UNIQUE (address, blockchain)
);

-- Table exchangepair holds all trading pairs for the pair scrapers.
-- The format has to be the same as emitted by the exchange's API in order
-- for the pair scrapers to be able to scrape trading data from the API.
CREATE TABLE exchangepair (
    exchangepair_id UUID DEFAULT gen_random_uuid(),
    symbol text not null,
    foreignname text not null,
    exchange text not null,
    UNIQUE (foreignname, exchange),
    -- These fields reference asset table and should be verified by pairdiscoveryservice.
    -- Only trades with verified pairs are processed further and thereby enter price calculation.
    verified boolean default false,
    id_quotetoken uuid REFERENCES asset(asset_id),
    id_basetoken uuid REFERENCES asset(asset_id)
);

CREATE TABLE exchangesymbol (
    exchangesymbol_id UUID DEFAULT gen_random_uuid(),
    symbol text not null,
    exchange text not null,
    UNIQUE (symbol,exchange),
    verified boolean default false,
    asset_id uuid REFERENCES asset(asset_id)
);

-- blockchain table stores all blockchains available in our databases
CREATE TABLE blockchain (
    blockchain_id integer primary key generated always as identity,
    UNIQUE name text not null,
    genesisdate timestamp,
    nativetoken text,
	verificationmechanism text
);


---------------------------------------
------- tables for NFT storage --------
---------------------------------------

-- nftclass is uniquely defined by the pair (blockchain,address),
-- referring to the blockchain on which the nft was minted.
CREATE TABLE nftclass (
    nftclass_id UUID DEFAULT gen_random_uuid(),
    address text not null,
    symbol text,
    name text,
    blockchain text REFERENCES blockchain(name),
    contract_type text,
    category text REFERENCES nftcategories(category),
    UNIQUE(blockchain,address)
);

-- an element from nft is a specific non-fungible nft, unqiuely
-- identified by the pair (address(on blockchain), tokenID)
CREATE TABLE nft (
    nft_id UUID DEFAULT gen_random_uuid(),
    nftclass_id uuid REFERENCES nftclass(nftclass_id),
    tokenID text not null,
    creation_time timestamp,
    creator_address text,
    attributes json,
    uri text
);

-- collect all possible categories for nfts
CREATE TABLE nftcategories (
    category_id UUID DEFAULT gen_random_uuid(),
    category text not null,
    UNIQUE(category)
);

CREATE TABLE nftsales (
    sale_id UUID DEFAULT gen_random_uuid(),
    nft_id uuid REFERENCES nft(nft_id),
    time timestamp,
    price_usd numeric,
    transfer_from text,
    transfer_to text,
    marketplace text
);

CREATE TABLE nftoffers (
    offer_id UUID DEFAULT gen_random_uuid(),
    nft_id uuid REFERENCES nft(nft_id),
    time timestamp,
    time_expiration timestamp,
    price_usd,
    from_address text,
    marketplace text
);