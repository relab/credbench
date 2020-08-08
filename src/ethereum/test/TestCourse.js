const { BN, expectEvent, constants, time, expectRevert } = require('@openzeppelin/test-helpers');
const { expect } = require('chai');

const Course = artifacts.require('CourseMock');

contract('Course', accounts => {
    const [teacher, evaluator, student, other] = accounts;
    let course = null;
    const digest1 = web3.utils.keccak256(web3.utils.toHex('cert1'));
    const digest2 = web3.utils.keccak256(web3.utils.toHex('cert2'));
    const digest3 = web3.utils.keccak256(web3.utils.toHex('cert3'));
    const courseDigest = web3.utils.keccak256(web3.utils.toHex('course1'));

    describe('constructor', () => {
        it('should successfully deploy the contract', async () => {
            course = await Course.new([teacher, evaluator], 2);
            (await course.isOwner(teacher)).should.equal(true);
            (await course.isOwner(evaluator)).should.equal(true);
            expect(await course.quorum()).to.be.bignumber.equal(new BN(2));
        });
    });

    describe('CRUD operations', () => {

        beforeEach(async () => {
            course = await Course.new([teacher, evaluator], 2);
            await time.increase(time.duration.seconds(1));
        });

        it('should add a new student', async () => {
            const { logs } = await course.addStudent(student, { from: teacher });

            expectEvent.inLogs(logs, 'StudentAdded', {
                student: student,
                requester: teacher
            });

            (await course.enrolledStudents(student)).should.equal(true);
        });

        it('only owner should be able to add a student', async () => {
            await expectRevert(
                course.addStudent(student, { from: other }),
                "Owners: sender is not an owner"
            );
        });

        it('should not add a student twice', async () => {
            await course.addStudent(student, { from: teacher });
            await expectRevert(
                course.addStudent(student, { from: evaluator }),
                "Course: student already registered in this course"
            );
        });

        it('should verify if a student is enrolled in the course', async () => {
            (await course.isEnrolled(student)).should.equal(false);
            await course.addStudent(student, { from: teacher });
            (await course.isEnrolled(student)).should.equal(true);
        });

        it('should not allow zero address', async () => {
            await expectRevert(
                course.addStudent(constants.ZERO_ADDRESS, { from: teacher }),
                "Course: student is the zero address"
            );
        });

        it('only owner should be able to remove an enrolled student', async () => {
            await course.addStudent(student, { from: evaluator });
            (await course.enrolledStudents(student)).should.equal(true);

            await expectRevert(
                course.removeStudent(student, { from: other }),
                'Owners: sender is not an owner'
            );
        });

        it('should allow an owner to remove an enrolled student', async () => {
            await course.addStudent(student, { from: evaluator });
            (await course.enrolledStudents(student)).should.equal(true);

            const { logs } = await course.removeStudent(student, { from: teacher });

            expectEvent.inLogs(logs, 'StudentRemoved', {
                student: student,
                requester: teacher
            });

            (await course.enrolledStudents(student)).should.equal(false);
        });

        it('should revert if try to remove an unregistered student', async () => {
            await expectRevert(
                course.removeStudent(student, { from: teacher }),
                "Course: student not registered"
            );
        });

        it('should allow a student to renounce the course', async () => {
            await course.addStudent(student, { from: teacher });
            (await course.enrolledStudents(student)).should.equal(true);

            const { logs } = await course.renounceCourse({ from: student });

            expectEvent.inLogs(logs, 'StudentRemoved', {
                student: student,
                requester: student
            });

            (await course.enrolledStudents(student)).should.equal(false);
        });
    });

    describe('Certification', () => {

        beforeEach(async () => {
            course = await Course.new([teacher, evaluator], 2);
            await time.increase(time.duration.seconds(1));
        });

        it('should issue a credential for a enrolled student', async () => {
            await course.addStudent(student, { from: teacher });
            await course.registerCredential(student, digest1, { from: teacher });
        });

        it('should not issue a credential for a non-enrolled address', async () => {
            await expectRevert(
                course.registerCredential(other, digest1, { from: teacher }),
                'Course: student not registered'
            );
        });

        it('should aggregate the course certificates of a subject', async () => {
            await course.enrollStudents([student]);

            for (d of [digest1, digest2, digest3]) {
                await course.registerCredential(student, d, { from: teacher });
                await course.registerCredential(student, d, { from: evaluator });
                await course.confirmCredential(d, { from: student });
                await time.increase(time.duration.seconds(1));

                (await course.certified(d)).should.equal(true);
            }

            await course.registerCredential(student, courseDigest, { from: teacher });
            await course.registerCredential(student, courseDigest, { from: evaluator });
            await course.confirmCredential(courseDigest, { from: student });

            // await time.increase(time.duration.hours(1));
            // (await course.hasEnded()).should.equal(true);

            const aggregated = await course.aggregateCredentials.call(student);
            let expected = web3.utils.keccak256(web3.eth.abi.encodeParameter('bytes32[]', [digest1, digest2, digest3, courseDigest]));
            (aggregated).should.equal(expected);
        });
    });

    describe('Aggregation', () => {
        const digests = [digest1, digest2, digest3, courseDigest];
        const expected = web3.utils.keccak256(web3.eth.abi.encodeParameter('bytes32[]', digests));

        beforeEach(async () => {
            course = await Course.new([teacher, evaluator], 2);
            await course.enrollStudents([student]);
            await time.increase(time.duration.seconds(1));

            for (d of digests) {
                await course.registerCredential(student, d, { from: teacher });
                await course.registerCredential(student, d, { from: evaluator });
                await course.confirmCredential(d, { from: student });
                await time.increase(time.duration.seconds(1));
            }
        });

        it('should not perform aggregation for unregistered students', async () => {
            // await time.increase(time.duration.hours(1));
            // (await course.hasEnded()).should.equal(true);

            await expectRevert(
                course.aggregateCredentials(other, { from: teacher }),
                'Course: student not registered'
            );
        });
    });
});
