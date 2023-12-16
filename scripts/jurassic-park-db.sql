CREATE TABLE `cage`
(
    `id` INT NOT NULL AUTO_INCREMENT,
    `externalId` VARCHAR(16) NOT NULL,
    `capacity` INT NOT NULL,
    `hasPower` TINYINT NOT NULL,
    `createdTime` DATETIME(6) DEFAULT NOW(6),

    PRIMARY KEY(`id`)
);
CREATE UNIQUE INDEX `cage_externalId` ON `cage`(`externalId`);

CREATE TABLE `speciesDiet`
(
    `name` VARCHAR(16) NOT NULL,
    PRIMARY KEY(`name`)
);
INSERT INTO `speciesDiet`(`name`)
VALUES('Carnivore'),
      ('Herbivore');


CREATE TABLE `species`
(
    `name` VARCHAR(16) NOT NULL,
    `diet` VARCHAR(16) NOT NULL,
    CONSTRAINT `species_diet` FOREIGN KEY(`diet`) REFERENCES `speciesDiet`(`name`),
    PRIMARY KEY(`name`)
);

INSERT INTO `species`(`name`, `diet`)
VALUES('Tyrannosaurus','Carnivore'),
      ('Velociraptor', 'Carnivore'),
      ('Spinosaurus', 'Carnivore'),
      ('Megalosaurus', 'Carnivore'),
      ('Brachiosaurus', 'Herbivore'),
      ('Stegosaurus', 'Herbivore'),
      ('Ankylosaurus', 'Herbivore'),
      ('Triceratops', 'Herbivore');


CREATE TABLE `sex`
(
    `name` VARCHAR(8) NOT NULL,
    PRIMARY KEY (`name`)
);

INSERT INTO `sex`(`name`)
VALUES('Female');

CREATE TABLE `dinosaur`
(
    `id` INT NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(16) NOT NULL,
    `species` VARCHAR(16) NOT NULL,
    `sex` VARCHAR(8) NOT NULL DEFAULT 'Female',
    `cageId` INT NULL,
    CONSTRAINT `dinosaur_species_fk` FOREIGN KEY(`species`) REFERENCES `species`(`name`),
    CONSTRAINT `dinosaur_cageId_fk` FOREIGN KEY(`cageId`) REFERENCES `cage`(`id`),
    CONSTRAINT `dinosaur_sex_fk` FOREIGN KEY(`sex`) REFERENCES `sex`(`name`),
    PRIMARY KEY(`id`)
);
CREATE UNIQUE INDEX `dinosaur_name` ON `dinosaur`(`name`);